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

// Node modules
var PHASER_PATH = './node_modules/phaser/build';
var BOOTSTRAP_PATH = './node_modules/bootstrap/dist';
var BOOTSTRAP_SOCIAL_PATH = './node_modules/bootstrap-social';
var FONT_AWESOME_PATH = './node_modules/font-awesome';
var PHASER_PLUGIN_SCENE_GRAPH = './node_modules/phaser-plugin-scene-graph/dist';

// App pathes
var SERVER_PATH = './../bbg_server';
var DJANGO_PATH = './';

// Static path
var STATIC_URL = 'www/bbgdev1.ga';
var STATIC_PATH = './static/bbg_client';
var BUILD_PATH = './../' + STATIC_URL;

var SCRIPTS_PATH = BUILD_PATH + '/build';
var CSS_PATH = SCRIPTS_PATH;

var SOURCE_PATH = './static/bbg_client/js';
var PROTO_PATH = './../protobufs';

var ENTRY_FILE = SOURCE_PATH + '/index.js';
var OUTPUT_FILE = 'game.js';

var keepFiles = false;

/**
 * Simple way to check for development/production mode.
 */
function isProduction() {
    return (typeof(argv.production) !== 'undefined') ? true : false;
}

function executeOnly() {
    return (typeof(argv.exec_only) !== 'undefined') ? true : false;
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
        del([BUILD_PATH.substr(2) + '/build/**/*.*']);
    } else {
        keepFiles = false;
    }
}

/**
 * Copies the content of the './protobufs'
 */
function copyProtobuf() {
    return gulp.src(PROTO_PATH + '/**/*')
        .pipe(gulp.dest(BUILD_PATH + PROTO_PATH.substr(4)));
}

/**
 * Update Go protobufs
 */
function updateGoProtobuf() {
    // golang
    exec('mkdir -p ' + SERVER_PATH.substr(2) + '/protobufs')
    // del([SERVER_PATH.substr(2) + '/protobufs/**/*.*']);
    exec('protoc ' + PROTO_PATH.substr(2) + '/*.proto ' +
         '--proto_path=' + PROTO_PATH.substr(2) + ' ' +
         '--go_out=' + SERVER_PATH.substr(2) + '/protobufs')

    // python
    exec('mkdir -p protobufs')
    // del(['protobufs/**/*.*']);
    exec('protoc ' + PROTO_PATH.substr(2) + '/*.proto ' +
         '--proto_path=' + PROTO_PATH.substr(2) + ' ' +
         '--python_out=protobufs')
}

function copyJS() {

    var srcList = [
        PHASER_PATH + '/phaser.min.js',
        BOOTSTRAP_PATH + '/js/bootstrap.min.js',
        PHASER_PLUGIN_SCENE_GRAPH + '/SceneGraph.js',
    ];
        
    return gulp.src(srcList)
        .pipe(gulp.dest(SCRIPTS_PATH));

}

function copyCSS() {

    var cssList = [
        BOOTSTRAP_PATH + '/css/bootstrap.min.css',
        BOOTSTRAP_SOCIAL_PATH + '/bootstrap-social.css',
        FONT_AWESOME_PATH + '/css/font-awesome.min.css',
        FONT_AWESOME_PATH + '/fonts/fontawesome-webfont.ttf',
        FONT_AWESOME_PATH + '/fonts/fontawesome-webfont.woff',
        FONT_AWESOME_PATH + '/fonts/fontawesome-webfont.woff2',
        FONT_AWESOME_PATH + '/fonts/fontawesome-webfont.eot',
        FONT_AWESOME_PATH + '/fonts/fontawesome-webfont.svg',
    ];

    return gulp.src(cssList)
        .pipe(gulp.dest(CSS_PATH));

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
            paths: [path.join(__dirname, 'static/bbg_client/js')],
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
        open: false,
        notify: false,
        port: 8001,
        localOnly: true,
        online: false,
    };
    
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

    if (executeOnly()) {
        return process.exit(0);
    }

}


gulp.task('cleanBuild', cleanBuild);
gulp.task('copyCSS', ['cleanBuild'], copyCSS);
gulp.task('copyJS', ['copyCSS'], copyJS);
gulp.task('copyProtobuf', ['copyJS'], copyProtobuf);
gulp.task('updateGoProtobuf', ['copyProtobuf'], updateGoProtobuf);
gulp.task('build', ['updateGoProtobuf'], build);
gulp.task('fastBuild', build);
gulp.task('serve', ['build'], serve);

if (!isProduction()) {
    gulp.task('watch-py', [], browserSync.reload);
    gulp.task('watch-js', ['fastBuild'], browserSync.reload);
    gulp.task('watch-static', ['copyJS', 'copyCSS'], browserSync.reload);
    gulp.task('watch-proto', ['copyProtobuf', 'updateGoProtobuf'], browserSync.reload);
}

gulp.task('default', ['serve']);
