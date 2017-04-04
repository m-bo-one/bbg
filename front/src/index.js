import MainState from 'states/MainState';

class Game extends Phaser.Game {

	constructor() {
		super(1024, 768, Phaser.CANVAS, 'content', null);
		this.state.add('MainState', MainState, false);
		this.state.start('MainState');
        this.DEBUG = true;
	}
}

window.game = new Game();
// let loggerF = window.console.log;
// window.console.log = function(...msg) {
//     if (window.game.DEBUG) {
//         loggerF(...msg);
//     }
// }
