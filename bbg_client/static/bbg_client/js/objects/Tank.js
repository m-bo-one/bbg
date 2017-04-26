import BaseElement from 'objects/BaseElement';
import HUD from 'objects/HUD';
import {pprint} from 'utils/helpers';

class Tank extends BaseElement {

    constructor(game, data, tankKey, turretKey) {
        super(game, data);
        let x = data.x, y = data.y;
        this.tankSprite = this.game.add.sprite(x, y, tankKey);
        this.tankSprite.anchor.setTo(0.5);
        this.tankSprite.scale.setTo(this.game.scaleRatio, this.game.scaleRatio);
        this.game.currentState.midLayer.add(this.tankSprite);

        // initialize turret sprite
        this.turretSprite = this.game.add.sprite(x, y, turretKey);
        this.turretSprite.anchor.setTo(0.25, 0.5);
        this.turretSprite.scale.setTo(this.game.scaleRatio, this.game.scaleRatio);
        this.game.currentState.midLayer.add(this.turretSprite);

        // add nickname
        this.textNick = this.game.add.text(x, y, data.name);
        this.textNick.scale.setTo(this.game.scaleRatio, this.game.scaleRatio);
        this.textNick.anchor.setTo(0.5, 2);
        this.game.currentState.midLayer.add(this.textNick);

        this.update(data);

        this.cmdId = 1;
        this.defFireRate = this.fireRate;

        this.eventType = this.game.stream.pbProtocol.Type;

        this.bullets = {};
    }

    getSprite() {
        return this.tankSprite;
    }

    destroy() {
        this.tankSprite.destroy();
        this.turretSprite.destroy();
        this.textNick.destroy();
    }

    update(data) {
        this.id = data.tankId;
        this.fireRate = data.fireRate;
        this.health = data.health;
        this.x = data.x;
        this.y = data.y;
        this.speed = data.speed;
        this.direction = data.direction;
        this.turretAngle = data.angle;
        this.damage = data.damage;
    }

    isDead() {
        return this.health <= 0;
    }

    changeColor(color=null) {
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
