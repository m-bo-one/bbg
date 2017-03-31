class Tank {

    constructor(game, data, tankKey, turretKey) {
        let x = data.x, y = data.y;
        this.game = game;
        this.tankSprite = this.game.add.sprite(x, y, tankKey);

        this.nextFire = 0;

        this.game.tanksGroup.add(this.tankSprite);

        this.tankSprite.scale.setTo(0.25, 0.25);
        this.tankSprite.anchor.setTo(0.5, 0.5);

        // initialize turret sprite
        this.turretSprite = this.game.add.sprite(x, y, turretKey);
        this.turretSprite.scale.setTo(0.25, 0.25);
        this.turretSprite.anchor.setTo(0.25, 0.5);

        // initialize bullets for tank
        this.bullets = this.game.add.group();
        this.bullets.enableBody = true;
        this.bullets.createMultiple(data.bullets, 'bullet', 0, false);
        this.bullets.setAll('scale.x', 0.25);
        this.bullets.setAll('scale.y', 0.25);
        this.bullets.setAll('anchor.x', 0.5); 
        this.bullets.setAll('anchor.y', 0.5);
        this.bullets.setAll('outOfBoundsKill', true);
        this.bullets.setAll('checkWorldBounds', true);

        this.d2a = {
            [this.game.stream.proto.Direction.N]: 360,
            [this.game.stream.proto.Direction.S]: 180,
            [this.game.stream.proto.Direction.E]: 90,
            [this.game.stream.proto.Direction.W]: -90,
        }

        this.update(data);

        this.cmdId = 1;
        this.defFireRate = this.fireRate;

        this.eventType = this.game.stream.pbProtocol.Type;
    }

    getSprite() {
        return this.tankSprite;
    }

    destroy() {
        this.bullets.callAll('kill');
        this.bullets.destroy();
        this.tankSprite.destroy();
        this.turretSprite.destroy();
    }

    update(data) {
        this.id = data.tankId;
        this.fireRate = data.fireRate;
        this.health = data.health;
        // this.bullets = data.bullets;
        this.x = data.x;
        this.y = data.y;
        this.speed = data.speed;
        this.direction = data.direction;
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
        if (this.game.time.now > this.nextFire && this.bullets.countDead() > 0) {
            this.nextFire = this.game.time.now + this.fireRate;

            let bullet = this.bullets.getFirstDead();

            bullet.reset(this.turretSprite.x, this.turretSprite.y);
            bullet.anchor.x = -5;
            bullet.rotation = this.game.physics.arcade.angleToPointer(this.turretSprite);
            this.game.physics.arcade.moveToPointer(bullet, 300, 0);
            if (this.fireRate > 30) {
                if (this.fireRate > this.defFireRate * 0.95) {
                    this.fireRate -= 1;
                } else if (this.fireRate > this.defFireRate * 0.9) {
                    this.fireRate -= 5;
                } else {
                    this.fireRate -= 15;
                }
            }
        } 
    }

    set rotation(angle) {
        this.turretSprite.rotation = angle;
    }

    rotate() {
        // console.log("My calc rotation: ", Math.atan2(this.game.input.mousePointer.worldY - this.turretSprite.y, this.game.input.mousePointer.worldX - this.turretSprite.x));
        this.rotation = this.game.physics.arcade.angleToPointer(this.turretSprite);
        console.log("Turret rotation: ", this.turretSprite.rotation);
        this.syncData('TankRotate', {
            x: this.x,
            y: this.y,
            mouseAxes: {
                x: this.game.input.mousePointer.worldX,
                y: this.game.input.mousePointer.worldY
            }
        });
    }

    set direction(direction) {
        if (this.game.stream.proto.Direction.hasOwnProperty(direction)) {
            direction = this.game.stream.proto.Direction[direction];
        }
        this.tankSprite.angle = this.d2a[direction];
    }

    move(direction) {
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
        return this.tankSprite.x;
    }

    set x(coord) {
        this.tankSprite.x = coord;
        this.turretSprite.x = coord;
    }

    get y() {
        return this.tankSprite.y;
    }

    set y(coord) {
        this.tankSprite.y = coord;
        this.turretSprite.y = coord;
    }

}

export default Tank;
