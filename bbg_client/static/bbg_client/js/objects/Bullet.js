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
        this.game.currentState.midLayer.add(this.bulletSprite);

        this.update(data);

        this.tank.bullets[this.id] = this;

        this._worker = setInterval(() => this.gcCleaner(), 1000);
    }

    static wsUpdate(game, data) {
        if (game.tanks.hasOwnProperty(data.tankId)) {
            let tank = game.tanks[data.tankId];
            let bullet;
            if (tank.bullets.hasOwnProperty(data.id)) {
                // console.log(`Update bullet position...`);
                tank.bullets[data.id].update(data);
            } else {
                // console.log(`Creating new bullet...`);
                new Bullet(game, data, 'bullet', tank);
            }
        }
    }

    gcCleaner() {
        if (!this.alive || Math.floor(Date.now() / 1000) > this.updatedAt + 2) {
            this.destroy();
        }
    }

    getSprite() {
        return this.bulletSprite;
    }

    destroy() {
        clearInterval(this._worker);
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
        this.updatedAt = Math.floor(Date.now() / 1000);

        if (!this.alive) {
            this.destroy();
        }
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
