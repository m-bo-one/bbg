import BaseElement from 'objects/BaseElement';

class Bullet extends BaseElement {

    constructor(game, data, key, tank) {
        super(game, data);
        let x = data.x, y = data.y;

        this.game = game;
        this.tank = tank;

        this.bulletSprite = this.game.add.sprite(x, y, key);
        this.bulletSprite.scale.setTo(0.25, 0.25);
        this.bulletSprite.anchor.setTo(0.5, 0.5);

        this.update(data);

        this.tank.bullets[this.id] = this;
    }

    getSprite() {
        return this.bulletSprite;
    }

    destroy() {
        this.bulletSprite.destroy();
        delete this.tank.bullets[this.id];
    }

    update(data) {
        this.id = data.id;
        this.x = data.x;
        this.y = data.y;
        this.angle = data.angle;
        this.speed = data.speed;
        this.tankId = data.tankId;
        this.alive = data.alive;

        if (!this.alive) {
            this.destroy();
        }

        // setTimeout(() => this.destroy(), 1000);
    }

    get angle() {
        return this.bulletSprite.rotation;
    }

    set angle(a) {
        this.bulletSprite.rotation = a;
    }

    set direction(direction) {
        if (this.game.stream.proto.Direction.hasOwnProperty(direction)) {
            direction = this.game.stream.proto.Direction[direction];
        }
        this.tankSprite.angle = this.d2a[direction];
    }

}

export default Bullet;
