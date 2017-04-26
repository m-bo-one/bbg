import Tank from 'objects/Tank';
import Bullet from 'objects/Bullet';
import HUD from 'objects/HUD';
import Player from 'objects/Player';
import ProtoStream from 'utils/ws';
import {pprint, isDeviceMobile, getKeyByValue, toFirstLowerCase} from 'utils/helpers';

class GameState extends Phaser.State {

    init(data) {
        this.initData = data;
    }

    preload() {
        this.game.imageLoad('tank', 'sprites/Tank2.png');
        this.game.imageLoad('gun-turret', 'sprites/Tank2GunTurret.png');
        this.game.imageLoad('bullet', 'sprites/Bullet.png');
        this.game.imageLoad('lazer', 'sprites/Lazer.png');

        this.game.tilemapLoad('map', 'tilemaps/map.json', null, Phaser.Tilemap.TILED_JSON);

        this.game.imageLoad('tiles', 'tilemaps/roads.png');

        this.game.scale.scaleMode = Phaser.ScaleManager.RESIZE;
        this.game.scale.pageAlignHorizontally = true;
        this.game.scale.pageAlignVertically = true;
        this.game.scale.forceLandscape = true;
        this.game.scale.parentIsWindow = true;
        this.game.scale.refresh();

        window.addEventListener('resize', () => {
            this.game.scale.refresh();
        });
    }

    create() {
        this.game.physics.startSystem(Phaser.Physics.ARCADE);
        this.stage.backgroundColor = "#ffffff";
        this.game.clearMenu();
        this.game.startGameSheet();

        // this.game.canvas.style.border = "1px solid black";
        this.game.tanks = {};

        this.backLayer = this.game.add.group();
        this.midLayer = this.game.add.group();
        this.frontLayer = this.game.add.group();

        this.createMap();

        this.tickRate = 120;

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
        });
    }

    createMap() {
        this.map = game.add.tilemap('map');
        this.map.addTilesetImage('spritesheet', 'tiles');

        let background = this.map.createLayer('Background', undefined, undefined, this.backLayer);
        background.resizeWorld();

        let road = this.map.createLayer('Road', undefined, undefined, this.backLayer);
        road.resizeWorld();

        // this.world.setBounds(0, 0, 2000, 2000);
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

    startHeartbeat() {
        this._shId = setInterval(() => {
            this.game.stream.send('Heartbeat', {});
        }, 1000 / this.tickRate);
    }

    stopHeartbeat() {
        clearInterval(this._shId);
        delete this._shId;
    }

    wsUpdate(data) {
        let stream = this.game.stream;
        let kData = getKeyByValue(stream.pbProtocol.Type, data.type);
        let pData = data[toFirstLowerCase(kData.slice(1))];

        if (typeof pData === 'undefined') return;

        switch(data.type) {
            case stream.pbProtocol.Type.TMapUpdate:
                switch(true) {
                    case Array.isArray(pData.tanks):
                        pData.tanks.forEach(dData => this.wsTankUpdate(dData));
                    case Array.isArray(pData.bullets):
                        pData.bullets.forEach(dData => this.wsBulletUpdate(dData));
                    case Array.isArray(pData.scores):
                        this.wsScoreUpdate(pData.scores);
                }
                break;
            case stream.pbProtocol.Type.TTankNew:
                this.wsPlayerCreate(pData);
                break;
            case stream.pbProtocol.Type.TTankRemove:
                this.wsTankRemove(pData);
                break;
            // case stream.pbProtocol.Type.TTankUpdate:
            //     this.wsTankUpdate(pData);
            //     break;
            // case stream.pbProtocol.Type.TBulletUpdate:
            //     this.wsBulletUpdate(pData);
            //     break;
            // case stream.pbProtocol.Type.TScoreUpdate:
            //     this.wsBulletUpdate(pData);
            //     break;
            // case stream.pbProtocol.Type.TPong:
            //     this.createOrUpdatePing(pData.timestamp - pData.processed);
            //     break;
            case stream.pbProtocol.Type.TUnhandledType:
                pprint('Unhandled type receive. Data: ', data);
                break;
        }
    }

    wsScoreUpdate(data) {
        this.game.player.get("HUD")
            .updateScore(data);
    }

    wsPlayerCreate(data) {
        if (!this.game.tanks.hasOwnProperty(data.tankId)) {
            pprint('Creating player...');

            this.game.player = new Player(game);
            this.game.player.add("HUD", new HUD(this.game, Object.assign(data, this.initData), this.frontLayer));
            this.game.player.add("tank", new Tank(this.game, data, 'tank', 'gun-turret'));

            this.game.currentTank = this.game.player.tank;

            this.game.camera.follow(this.game.currentTank.getSprite());
            this.game.input.addMoveCallback(
                this.game.currentTank.rotate.bind(this.game.currentTank));

            this.game.tanks[data.tankId] = this.game.currentTank;
            this.startHeartbeat();
        }
    }

    wsTankUpdate(data) {
        pprint('Receive tank update. Applying...');
        if (!this.game.tanks.hasOwnProperty(data.tankId)) {
            console.log('Creating new tank...');
            this.game.tanks[data.tankId] = new Tank(this.game, data, 'tank', 'gun-turret');
        } else if (this.game.player.tank.id == data.tankId) {
            this.game.player.update(data);
        } else {
            this.game.tanks[data.tankId].update(data);
        }
    }

    wsTankRemove(data) {
        if (this.game.tanks.hasOwnProperty(data.tankId)) {
            pprint(`Removing tank ID:${data.tankId}...`);

            if (this.game.currentTank && this.game.currentTank.id == data.tankId) {
                let keyboard = this.game.input.keyboard;
                keyboard.onDownCallback = keyboard.onUpCallback = keyboard.onPressCallback = null;
                this.game.input.moveCallbacks = [];
            } else {
                this.game.tanks[data.tankId].destroy();
                delete this.game.tanks[data.tankId];
            }
        }
    }

    wsBulletUpdate(data) {
        if (this.game.tanks.hasOwnProperty(data.tankId)) {
            let tank = this.game.tanks[data.tankId];
            let bullet;
            if (tank.bullets.hasOwnProperty(data.id)) {
                // console.log(`Update bullet position...`);
                tank.bullets[data.id].update(data);
            } else {
                // console.log(`Creating new bullet...`);
                new Bullet(this.game, data, 'bullet', tank);
            }
        }
    }

}

export default GameState;
