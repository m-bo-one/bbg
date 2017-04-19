import Tank from 'objects/Tank';
import Bullet from 'objects/Bullet';
import ProtoStream from 'utils/ws';
import {pprint, isDeviceMobile, getKeyByValue, toFirstLowerCase} from 'utils/helpers';

class GameState extends Phaser.State {

    init(data) {
        this.initData = data;
    }

    preload() {
        this.game.imageLoad('tank', 'sprites/Tank.png');
        this.game.imageLoad('gun-turret', 'sprites/GunTurret.png');
        this.game.imageLoad('bullet', 'sprites/Bullet.png');
        this.game.imageLoad('lazer', 'sprites/Lazer.png');
    }

    create() {
        this.game.physics.startSystem(Phaser.Physics.ARCADE);
        this.stage.backgroundColor = "#ffffff";
        this.game.clearMenu();
        this.game.startGameSheet();

        this.game.canvas.style.border = "1px solid black";
        this.game.tanks = {};

        this.game.backLayer = this.game.add.group();
        this.game.midLayer = this.game.add.group();
        this.game.frontLayer = this.game.add.group();

        this.cursors = this.game.input.keyboard.addKeys({
            W: Phaser.Keyboard.W,
            A: Phaser.Keyboard.A,
            S: Phaser.Keyboard.S,
            D: Phaser.Keyboard.D,
            SPACEBAR: Phaser.Keyboard.SPACEBAR
        });
        this.game.time.advancedTiming = true; // enable FPS
        this.game.stream = new ProtoStream(`ws://${predefinedVars.wsURL}/game`, () => {
            this.game.stream.send("TankReg", {
                token: predefinedVars.currentUser.token,
                tKey: this.initData.tkey
            });
            let callbackType = isDeviceMobile() ? "pagehide" : "beforeunload";
            let callback = (e) => {
                window.removeEventListener(callbackType, callback);
                this.game.stream.send('TankUnreg');
            };
            window.addEventListener(callbackType, callback);

            this.createStatBlock();
        });
    }

    createStatBlock() {
        let initX = 30;
        let initY = 20;
        let offset = 30;
        this._scoreText = this.game.add.text(initX, initY, `Scores: ${this.initData["scores-count"]}`);
        this._killText = this.game.add.text(initX, initY + offset, `Kills: ${this.initData["kill-count"]}`);
        this._deathText = this.game.add.text(initX, initY + 2 * offset, `Death: ${this.initData["death-count"]}`);
    }

    createRespawnBlock(counter=3) {
        if (typeof this._restartText === "object") return;
        this._restartText = this.game.add.text(0, 0, `Respawn at: ${counter}`);
        this._restartText.x = this.game.width - 50 - this._restartText.width;
        this._restartText.y = this.game.height - 75;
        let id = setInterval(() => {
            counter--;
            if (counter === 0) {
                clearInterval(id);
                this._restartText.destroy();
                delete this._restartText;
                return
            }
            this._restartText.text = `Respawn at: ${counter}`;
        }, 1000);
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
        if (predefinedVars.debug) {
            if (!this.game.currentTank) return;
        }
        // if (!this.game.currentTank && !predefinedVars.debug) return;
        // this.game.debug.text(`FPS: ${this.game.time.fps}`, 2, 14, "#00ff00");
        // this.game.debug.text(`HP: ${this.game.currentTank.health}`, 2, 14 * 2, "#00ff00");
        // this.game.debug.spriteInfo(this.game.currentTank.getSprite(), 640, 14);
    }

    wsUpdate(data) {
        let stream = this.game.stream;
        let kData = getKeyByValue(stream.pbProtocol.Type, data.type);
        let pData = data[toFirstLowerCase(kData)];

        if (typeof pData === 'undefined') return;

        pprint('Received message: ', pData)

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
                pprint('Unhandled type receive. Data: ', data);
                break;
        }
    }

}

export default GameState;
