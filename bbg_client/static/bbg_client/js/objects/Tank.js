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

    static wsCreate(game, data) {
        if (!game.tanks.hasOwnProperty(data.tankId)) {
            console.log('Creating new tank...');
            game.tanks[data.tankId] = new Tank(game, data, 'tank', 'gun-turret');
            game.currentTank = game.tanks[data.tankId];

            let callback = game.currentTank.rotate.bind(game.currentTank);

            game.input.addMoveCallback(callback);
        }
    }

    static wsUpdate(game, data) {
        console.log('Receive tank update. Applying...');
        if (!game.tanks.hasOwnProperty(data.tankId)) {
            console.log('Creating new tank...');
            game.tanks[data.tankId] = new Tank(game, data, 'tank', 'gun-turret');
        } else {
            game.tanks[data.tankId].update(data);
        }
    }

    static wsRemove(game, data) {
        if (game.tanks.hasOwnProperty(data.tankId)) {
            console.log(`Removing tank ID:${data.tankId}...`);

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
    }

    set y(coord) {
        super.y = coord;
        this.turretSprite.y = coord;
    }

}

export default Tank;
