import BaseElement from 'objects/BaseElement';
import HealthBar from 'objects/HealthBar';
import {pprint} from 'utils/helpers';

class Tank extends BaseElement {

    constructor(game, data, tankKey, turretKey) {
        super(game, data);
        let x = data.x, y = data.y;
        this.tankSprite = this.game.add.sprite(x, y, tankKey);
        this.tankSprite.anchor.setTo(0.5);
        this.game.currentState.midLayer.add(this.tankSprite);

        // initialize turret sprite
        this.turretSprite = this.game.add.sprite(x, y, turretKey);
        this.turretSprite.anchor.setTo(0.25, 0.5);
        this.game.currentState.midLayer.add(this.turretSprite);

        // add nickname
        this.textNick = this.game.add.text(x, y, data.name);
        this.textNick.scale.setTo(0.5);
        this.textNick.anchor.setTo(0.5, 2);
        this.game.currentState.midLayer.add(this.textNick);

        this.update(data);

        this.cmdId = 1;
        this.defFireRate = this.fireRate;

        this.eventType = this.game.stream.pbProtocol.Type;

        this.bullets = {};
    }

    static wsCreate(game, data) {
        if (!game.tanks.hasOwnProperty(data.tankId)) {
            pprint('Creating new tank...');
            game.tanks[data.tankId] = new Tank(game, data, 'tank', 'gun-turret');
            game.currentTank = game.tanks[data.tankId];

            game.camera.follow(game.currentTank.getSprite());

            let healthBar = new HealthBar(game, {
                width: 100 * 2,
                height: 20,
                bar: {
                  color: 'red'
                },
                animationDuration: 500,
            });
            healthBar.setPosition(130, game.height - 50)
            healthBar.setWidth(data.health * 2);
            healthBar.setFixedToCamera(true);

            game.currentState.frontLayer.add(healthBar.bgSprite);
            game.currentState.frontLayer.add(healthBar.barSprite);
            game.currentTank.healthBar = healthBar;

            let callback = game.currentTank.rotate.bind(game.currentTank);

            game.input.addMoveCallback(callback);
        }
    }

    static wsUpdate(game, data) {
        pprint('Receive tank update. Applying...');
        if (!game.tanks.hasOwnProperty(data.tankId)) {
            pprint('Creating new tank...');
            game.tanks[data.tankId] = new Tank(game, data, 'tank', 'gun-turret');
        } else {
            game.tanks[data.tankId].update(data);
        }
    }

    static wsRemove(game, data) {
        if (game.tanks.hasOwnProperty(data.tankId)) {
            pprint(`Removing tank ID:${data.tankId}...`);

            if (game.currentTank && game.currentTank.id == data.tankId) {
                // game.currentTank.destroy();
                // delete game.currentTank;
                let keyboard = game.input.keyboard;
                keyboard.onDownCallback = keyboard.onUpCallback = keyboard.onPressCallback = null;
                game.input.moveCallbacks = [];
            } else {
                game.tanks[data.tankId].destroy();
                delete game.tanks[data.tankId];
            }
        }
    }

    getSprite() {
        return this.tankSprite;
    }

    destroy() {
        this.tankSprite.destroy();
        this.turretSprite.destroy();
        this.textNick.destroy();
        if (this === this.game.currentTank) this.healthBar.kill();
    }

    update(data) {
        if (this === this.game.currentTank && data.health != this.health)
            this.healthBar.setWidth(data.health * 2);

        this.id = data.tankId;
        this.fireRate = data.fireRate;
        this.health = data.health;
        this.x = data.x;
        this.y = data.y;
        this.speed = data.speed;
        this.direction = data.direction;
        this.turretAngle = data.angle;
        this.damage = data.damage;

        if (this.isDead()) {
            this.changeColor(0xff9a22);
            if (this === this.game.currentTank)
                this.game.currentState.createRespawnBlock();
        } else {
            // HARDCODED VALUE OF COLOR
            this.changeColor(16777215);
        }
    }

    isDead() {
        return this.health <= 0;
    }

    changeColor(color) {
        this._lastColor = this.tankSprite.tint;
        this.tankSprite.tint = color;
        this.turretSprite.tint = color;
    }

    syncData(type, data) {
        if (data) {
            data['id'] = this.cmdId;
            if (this.cmdId > 1) {
                data['prevId'] = this.cmdId - 1;
            }
            this.cmdId++;
            this.game.stream.send(type, data);
        } else {
            this.game.stream.send(type);
        }
    }

    fire() {
        if (this.isDead()) return;
        this.syncData('TankShoot', {
            x: this.x,
            y: this.y,
            mouseAxes: {
                x: this.game.input.mousePointer.worldX,
                y: this.game.input.mousePointer.worldY
            }
        });
    }

    set turretAngle(angle) {
        this.turretSprite.rotation = angle;
    }

    rotate() {
        if (this.isDead() || !this.game.input.mousePointer.withinGame) return;
        this.turretAngle = this.game.physics.arcade.angleToPointer(this.turretSprite);
        this.syncData('TankRotate', {
            x: this.x,
            y: this.y,
            mouseAxes: {
                x: this.game.input.mousePointer.worldX,
                y: this.game.input.mousePointer.worldY
            }
        });
    }

    move(direction) {
        if (this.isDead()) return;
        this.rotate();
        switch(direction) {
            case 'N':
                this.y -= this.speed;
                break;
            case 'S':
                this.y += this.speed;
                break;
            case 'E':
                this.x += this.speed;
                break;
            case 'W':
                this.x -= this.speed;
                break;
            default:
                return;
        }
        this.direction = direction;

        if (this.game.currentTank == this) {
            this.game.camera.focusOnXY(this.x, this.y);
        }
        this.syncData('TankMove', {
            direction: this.protoDirection[direction]
        });
    }

    stop() {
        // this.game.stream.send('haha');
    }

    get x() {
        return super.x;
    }

    get y() {
        return super.y;
    }

    set x(coord) {
        super.x = coord;
        this.turretSprite.x = coord;
        this.textNick.x = coord;
    }

    set y(coord) {
        super.y = coord;
        this.turretSprite.y = coord;
        this.textNick.y = coord;
    }

}

export default Tank;
