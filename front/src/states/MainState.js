import Tank from 'objects/Tank';
import ProtoStream from 'utils/ws';
import * as helpers from 'utils/helpers';

class MainState extends Phaser.State {

    preload() {
        this.game.load.image('tank', 'assets/sprites/Tank.png');
        this.game.load.image('gun-turret', 'assets/sprites/GunTurret.png');
        this.game.load.image('bullet', 'assets/sprites/Bullet.png');
    }

    create() {
        this.game.physics.startSystem(Phaser.Physics.ARCADE);
        this.stage.backgroundColor = "#DDDDDD";
        this.tanks = {};
        this.game.tanksGroup = this.game.add.group();

        this.cursors = this.game.input.keyboard.addKeys({
            W: Phaser.Keyboard.W,
            A: Phaser.Keyboard.A,
            S: Phaser.Keyboard.S,
            D: Phaser.Keyboard.D,
            SPACEBAR: Phaser.Keyboard.SPACEBAR
        });

        this.game.stream = new ProtoStream('ws://localhost:8080/game');
        this.game.stream.onLoadComplete(() => {
            this.game.stream.send("TankReg");
            window.onbeforeunload = () => this.game.stream.send('TankUnreg');
        });
    }

    update() {
        if (!this.currentTank) return;
        if (this.cursors.SPACEBAR.isDown) this.currentTank.fire()  
        switch(true) {
            case this.cursors.W.isDown:
                this.currentTank.move('N');
                break;
            case this.cursors.S.isDown:
                this.currentTank.move('S');
                break;
            case this.cursors.D.isDown:
                this.currentTank.move('E');
                break;
            case this.cursors.A.isDown:
                this.currentTank.move('W');
                break;
            default:
                this.currentTank.stop();
        }
    }

    render() {
        if (!this.currentTank) return;
        this.game.debug.text(this.game.time.fps || '--', 2, 14, "#00ff00");
        this.game.debug.text(`HP: ${this.currentTank.health}`, 2, 14 * 2, "#00ff00");
        this.game.debug.text(`Fire Rate: ${this.currentTank.fireRate}`, 2, 14 * 3, "#00ff00");
        this.game.debug.spriteInfo(this.currentTank.getSprite(), 640, 14);
    }

    wsUpdate(data) {
        let stream = this.game.stream;
        let kData = helpers.getKeyByValue(stream.pbProtocol.Type, data.type);
        let pData = data[helpers.toFirstLowerCase(kData)];
        if (typeof pData === 'undefined') return;

        switch(data.type) {
            case stream.pbProtocol.Type.TankNew:
                if (!this.tanks.hasOwnProperty(pData.tankId)) {
                    console.log('Creating new tank...');
                    this.tanks[pData.tankId] = new Tank(this.game, pData, 'tank', 'gun-turret');
                    this.currentTank = this.tanks[pData.tankId];
                    this.game.input.keyboard.onUpCallback = (e) => {
                        this.currentTank.fireRate = this.currentTank.defFireRate;
                    };
                    let callback = this.currentTank.rotate.bind(this.currentTank);
                    this.game.input.addMoveCallback(callback);
                }
                break;
            case stream.pbProtocol.Type.TankUpdate:
                console.log('Receive tank update. Applying...');
                if (!this.tanks.hasOwnProperty(pData.tankId)) {
                    console.log('Creating new tank...');
                    this.tanks[pData.tankId] = new Tank(this.game, pData, 'tank', 'gun-turret');
                } else {
                    this.tanks[pData.tankId].update(pData);
                }
                break;
            case stream.pbProtocol.Type.TankRemove:
                if (this.tanks.hasOwnProperty(pData.tankId)) {
                    console.log(`Removing tank ID:${pData.tankId}...`);
                    this.tanks[pData.tankId].destroy();
                    delete this.tanks[pData.tankId];

                    if (this.currentTank.id == pData.tankId) {
                        this.currentTank.destroy();
                        delete this.currentTank;
                        let keyboard = this.game.input.keyboard;
                        keyboard.onDownCallback = keyboard.onUpCallback = keyboard.onPressCallback = null;
                    }
                }
                break;
            case stream.pbProtocol.Type.UnhandledType:
                console.log('Unhandled type receive. Data: ', data);
                break;
        }
    }

}

export default MainState;
