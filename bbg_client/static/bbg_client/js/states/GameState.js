import Tank from 'objects/Tank';
import Bullet from 'objects/Bullet';
import ProtoStream from 'utils/ws';
import * as helpers from 'utils/helpers';

class GameState extends Phaser.State {

    preload() {
        this.game.imageLoad('tank', 'sprites/Tank.png');
        this.game.imageLoad('gun-turret', 'sprites/GunTurret.png');
        this.game.imageLoad('bullet', 'sprites/Bullet.png');
        this.game.imageLoad('lazer', 'sprites/Lazer.png');
    }

    create() {
        this.game.physics.startSystem(Phaser.Physics.ARCADE);
        this.stage.backgroundColor = "#ffffff";
        this.game.canvas.style.border = "1px solid black";
        this.game.tanks = {};
        this.game.tanksGroup = this.game.add.group();

        this.cursors = this.game.input.keyboard.addKeys({
            W: Phaser.Keyboard.W,
            A: Phaser.Keyboard.A,
            S: Phaser.Keyboard.S,
            D: Phaser.Keyboard.D,
            SPACEBAR: Phaser.Keyboard.SPACEBAR
        });
        this.game.time.advancedTiming = true; // enable FPS
        this.game.stream = new ProtoStream(`ws://${predefinedVars.wsURL}/game`);
        this.game.stream.onLoadComplete(() => {
            this.game.stream.send("TankReg", {
                token: predefinedVars.currentUser.token,
                tankId: 5
            });
            let callbackType = helpers.isDeviceMobile() ? "pagehide" : "beforeunload";
            let callback = (e) => {
                window.removeEventListener(callbackType, callback);
                this.game.stream.send('TankUnreg');
            };
            window.addEventListener(callbackType, callback);
        });
    }

    update() {
        if (!this.game.currentTank) return;
        if (this.cursors.SPACEBAR.isDown) this.game.currentTank.fire();
        switch(true) {
            case this.cursors.W.isDown:
                this.game.currentTank.move('N');
                break;
            case this.cursors.S.isDown:
                this.game.currentTank.move('S');
                break;
            case this.cursors.D.isDown:
                this.game.currentTank.move('E');
                break;
            case this.cursors.A.isDown:
                this.game.currentTank.move('W');
                break;
            default:
                this.game.currentTank.stop();
        }
    }

    render() {
        if (!this.game.currentTank) return;
        this.game.debug.text(`FPS: ${this.game.time.fps}`, 2, 14, "#00ff00");
        this.game.debug.text(`HP: ${this.game.currentTank.health}`, 2, 14 * 2, "#00ff00");
        this.game.debug.spriteInfo(this.game.currentTank.getSprite(), 640, 14);
    }

    wsUpdate(data) {
        let stream = this.game.stream;
        let kData = helpers.getKeyByValue(stream.pbProtocol.Type, data.type);
        let pData = data[helpers.toFirstLowerCase(kData)];

        if (typeof pData === 'undefined') return;

        console.log('Received message: ', pData)

        switch(data.type) {
            case stream.pbProtocol.Type.MapUpdate:
                switch(true) {
                    case Array.isArray(pData.tanks):
                        pData.tanks.forEach(dData => Tank.wsUpdate(game, dData));
                    case Array.isArray(pData.bullets):
                        pData.bullets.forEach(dData => Bullet.wsUpdate(game, dData));
                }
                break;
            case stream.pbProtocol.Type.TankNew:
                Tank.wsCreate(game, pData);
                break;
            case stream.pbProtocol.Type.TankUpdate:
                Tank.wsUpdate(game, pData);
                break;
            case stream.pbProtocol.Type.TankRemove:
                Tank.wsRemove(game, pData);
                break;
            case stream.pbProtocol.Type.BulletUpdate:
                Bullet.wsUpdate(game, pData);
                break;
            case stream.pbProtocol.Type.UnhandledType:
                console.log('Unhandled type receive. Data: ', data);
                break;
        }
    }

}

export default GameState;
