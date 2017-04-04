class BaseElement {

    constructor(game, data) {
        this.game = game;

        this.d2a = {
            [this.game.stream.proto.Direction.N]: 360,
            [this.game.stream.proto.Direction.S]: 180,
            [this.game.stream.proto.Direction.E]: 90,
            [this.game.stream.proto.Direction.W]: -90,
        }
    }

    update(data) {

    }

    getSprite() {

    }

    get x() {
        return this.getSprite().x;
    }

    set x(coord) {
        this._x = coord;
        this.getSprite().x = coord;
    }

    get y() {
        return this.getSprite().y;
    }

    set y(coord) {
        this._y = coord;
        this.getSprite().y = coord;
    }

    set direction(direction) {
        if (this.game.stream.proto.Direction.hasOwnProperty(direction)) {
            direction = this.game.stream.proto.Direction[direction];
        }
        this.getSprite().angle = this.d2a[direction];
    }
}

export default BaseElement;
