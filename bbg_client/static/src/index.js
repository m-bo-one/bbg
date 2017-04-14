import GameState from 'states/GameState';
import MainState from 'states/MainState';
import MenuState from 'states/MenuState';

class Game extends Phaser.Game {

    constructor() {
        super(1024, 768, Phaser.CANVAS, 'content', null);
        this.state.add('MainState', MainState, false);
        this.state.add('MenuState', MenuState, false);
        this.state.add('GameState', GameState, false);

        (predefinedVars.currentUser !== null) ? this.state.start('MenuState') : this.state.start('MainState');
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

    imageLoad(key, fileName) {
        this.load.image(key, predefinedVars.staticURL + fileName);
    }
}

window.game = new Game();
