var del = require('del');
var gulp = require('gulp');
var path = require('path');
var argv = require('yargs').argv;
var gutil = require('gulp-util');
var source = require('vinyl-source-stream');
var buffer = require('gulp-buffer');
var uglify = require('gulp-uglify');
var gulpif = require('gulp-if');
var exorcist = require('exorcist');
var babelify = require('babelify');
var browserify = require('browserify');
var browserSync = require('browser-sync');
var exec = require('child_process').exec;

/**
 * Using different folders/file names? Change these constants:
 */
var SERVER_PATH = './../server';
var PHASER_PATH = './node_modules/phaser/build/';
var BUILD_PATH = './build';
var SCRIPTS_PATH = './static/build/scripts';
var SOURCE_PATH = './static/src';
var STATIC_PATH = './static';
var PROTO_PATH = './static/protobufs';
var ENTRY_FILE = SOURCE_PATH + '/index.js';
var OUTPUT_FILE = 'game.js';

var keepFiles = false;

/**
 * Simple way to check for development/production mode.
 */
function isProduction() {
    return (typeof(argv.production) !== 'undefined') ? true : false;
}

/**
 * Logs the current build mode on the console.
 */
function logBuildMode() {
    
    if (isProduction()) {
        gutil.log(gutil.colors.green('Running production build...'));
    } else {
        gutil.log(gutil.colors.yellow('Running development build...'));
    }

}

/**
 * Deletes all content inside the './build' folder.
 * If 'keepFiles' is true, no files will be deleted. This is a dirty workaround since we can't have
 * optional task dependencies :(
 * Note: keepFiles is set to true by gulp.watch (see serve()) and reseted here to avoid conflicts.
 */
function cleanBuild() {
    if (!keepFiles) {
        del(['build/**/*.*']);
    } else {
        keepFiles = false;
    }
}

/**
 * Copies the content of the './static' folder into the '/build' folder.
 * Check out README.md for more info on the '/static' folder.
 */
function copyStatic() {
    return gulp.src(STATIC_PATH + '/**/*')
        .pipe(gulp.dest(BUILD_PATH));
}

/**
 * Copies the content of the './protobufs'
 */
function copyProtobuf() {
    return gulp.src(PROTO_PATH + '/**/*')
        .pipe(gulp.dest(BUILD_PATH + PROTO_PATH.substr(1)));
}

/**
 * Update Go protobufs
 */
function updateGoProtobuf() {
    del([SERVER_PATH + PROTO_PATH.substr(1) + '/**/*.*']);
    return exec('protoc protobufs/*.proto --go_out=../server');
}

/**
 * Copies required Phaser files from the './node_modules/Phaser' folder into the './build/scripts' folder.
 * This way you can call 'npm update', get the lastest Phaser version and use it on your project with ease.
 */
function copyPhaser() {

    var srcList = ['phaser.min.js'];
    
    if (!isProduction()) {
        srcList.push('phaser.map', 'phaser.js');
    }
    
    srcList = srcList.map(function(file) {
        return PHASER_PATH + file;
    });
        
    return gulp.src(srcList)
        .pipe(gulp.dest(SCRIPTS_PATH));

}

/**
 * Transforms ES2015 code into ES5 code.
 * Optionally: Creates a sourcemap file 'game.js.map' for debugging.
 * 
 * In order to avoid copying Phaser and Static files on each build,
 * I've abstracted the build logic into a separate function. This way
 * two different tasks (build and fastBuild) can use the same logic
 * but have different task dependencies.
 */
function build() {

    var sourcemapPath = SCRIPTS_PATH + '/' + OUTPUT_FILE + '.map';
    logBuildMode();

    return browserify({
            paths: [path.join(__dirname, 'static/src')],
            entries: ENTRY_FILE,
            debug: !isProduction(),
            transform: [
                [
                    babelify, {
                        presets: ["es2015"]
                    }
                ]
            ]
        })
        .transform(babelify)
        .bundle().on('error', function(error) {
            gutil.log(gutil.colors.red('[Build Error]', error.message));
            this.emit('end');
        })
        .pipe(gulpif(!isProduction(), exorcist(sourcemapPath)))
        .pipe(source(OUTPUT_FILE))
        .pipe(buffer())
        .pipe(gulpif(isProduction(), uglify()))
        .pipe(gulp.dest(SCRIPTS_PATH));

}

/**
 * Starts the Browsersync server.
 * Watches for file changes in the 'src' folder.
 */
function serve() {

    if (isProduction()) return;
    
    var options = {
        ui: false,
        proxy: '127.0.0.1:8000',
        open: "local",
        notify: false,
        port: 8001,
        localOnly: true,
        online: false,
    };

    // if (!isProduction()) {
    //     options.middleware = [
    //         function(req, res, next) {
    //             se = req.url.match(/\/login\/(.*)/)
    //             if (se && se.length > 1) {
    //                 res.writeHead(302, {'Location': 'http://127.0.0.1:8888' + req.url});
    //                 res.end();
    //             } else {
    //                 next();
    //             }
    //         }
    //     ];
    // }
    
    browserSync.init(options);

    gulp.watch('./**/*.py', ['watch-py']);

    // Watches for changes in files inside the './src' folder.
    gulp.watch(SOURCE_PATH + '/**/*.js', ['watch-js']);

    // Watches for changes in files inside the './protobufs' folder.
    gulp.watch(PROTO_PATH + '/**/*.proto', ['watch-proto']);
    
    // Watches for changes in files inside the './static' folder. Also sets 'keepFiles' to true (see cleanBuild()).
    gulp.watch(STATIC_PATH + '/**/*', ['watch-static']).on('change', function() {
        keepFiles = true;
    });

}


gulp.task('cleanBuild', cleanBuild);
gulp.task('copyStatic', ['cleanBuild'], copyStatic);
gulp.task('copyPhaser', ['copyStatic'], copyPhaser);
// gulp.task('copyProtobuf', ['copyPhaser'], copyProtobuf);
gulp.task('updateGoProtobuf', ['copyPhaser'], updateGoProtobuf);
gulp.task('build', ['updateGoProtobuf'], build);
gulp.task('fastBuild', build);
gulp.task('serve', ['build'], serve);

if (!isProduction()) {
    gulp.task('watch-py', [], browserSync.reload);
    gulp.task('watch-js', ['fastBuild'], browserSync.reload);
    gulp.task('watch-static', ['copyPhaser'], browserSync.reload);
    gulp.task('watch-proto', ['copyProtobuf', 'updateGoProtobuf'], browserSync.reload);
}

/**
 * The tasks are executed in the following order:
 * 'cleanBuild' -> 'copyStatic' -> 'copyPhaser' -> 'build' -> 'serve'
 * 
 * Read more about task dependencies in Gulp: 
 * https://medium.com/@dave_lunny/task-dependencies-in-gulp-b885c1ab48f0
 */
gulp.task('default', ['serve']);
