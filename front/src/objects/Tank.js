import BaseElement from 'objects/BaseElement';

class Tank extends BaseElement {

    constructor(game, data, tankKey, turretKey) {
        super(game, data);
        let x = data.x, y = data.y;
        this.tankSprite = this.game.add.sprite(x, y, tankKey);

        this.nextFire = 0;

        this.game.tanksGroup.add(this.tankSprite);

        this.tankSprite.scale.setTo(0.25, 0.25);
        this.tankSprite.anchor.setTo(0.5, 0.5);

        // initialize turret sprite
        this.turretSprite = this.game.add.sprite(x, y, turretKey);
        this.turretSprite.scale.setTo(0.25, 0.25);
        this.turretSprite.anchor.setTo(0.25, 0.5);

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
        // if (this.game.time.now > this.nextFire && this.bullets.countDead() > 0) {
        //     this.nextFire = this.game.time.now + this.fireRate;

        //     let bullet = this.bullets.getFirstDead();

        //     bullet.reset(this.turretSprite.x, this.turretSprite.y);
        //     bullet.anchor.x = -5;
        //     bullet.rotation = this.game.physics.arcade.moveToPointer(bullet, 1000);
        //     // bullet.body.velocity.x = 1000;
        //     // bullet.body.velocity.y = 1000;

        //     if (this.fireRate > 30) {
        //         if (this.fireRate > this.defFireRate * 0.95) {
        //             this.fireRate -= 1;
        //         } else if (this.fireRate > this.defFireRate * 0.9) {
        //             this.fireRate -= 5;
        //         } else {
        //             this.fireRate -= 15;
        //         }
        //     }
        // }
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
        if (!this.game.input.mousePointer.withinGame) return;
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

        this.syncData('TankMove', {
            direction: this.game.stream.proto.Direction[direction]
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
    }

    set y(coord) {
        super.y = coord;
        this.turretSprite.y = coord;
    }

}

export default Tank;
