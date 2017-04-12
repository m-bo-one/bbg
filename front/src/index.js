import GameState from 'states/GameState';
import MainState from 'states/MainState';

class Game extends Phaser.Game {

	constructor() {
		super(1024, 768, Phaser.CANVAS, 'content', null);
        this.DEBUG = true;
        this.state.add('MainState', MainState, false);
		this.state.add('GameState', GameState, false);
        this.state.start('MainState');
		// this.state.start('GameState');
	}

    create() {
        this.world.setBounds(0, 0, 2000, 2000);
    }

    flushMap() {
        for (let tkey in this.tanks) {
            let tank = this.tanks[tkey];
            for (let bkey in tank.bullets) {
                tank.bullets[bkey].destroy();
            }
            tank.bullets = {};
            tank.destroy();
        }
        this.tanks = {};

        delete this.currentTank;
        let keyboard = this.input.keyboard;
        keyboard.onDownCallback = keyboard.onUpCallback = keyboard.onPressCallback = null;
        game.input.moveCallbacks = [];
    }
}

window.game = new Game();
// let loggerF = window.console.log;
// window.console.log = function(...msg) {
//     if (window.game.DEBUG) {
//         loggerF(...msg);
//     }
// }
