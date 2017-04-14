class BaseElement {

    constructor(game, data) {
        this.game = game;

        this.d2a = {
            [this.protoDirection.N]: 360,
            [this.protoDirection.S]: 180,
            [this.protoDirection.E]: 90,
            [this.protoDirection.W]: -90,
        }
    }

    get protoDirection() {
        return this.game.stream.proto.Direction.values;
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
        if (this.protoDirection.hasOwnProperty(direction)) {
            direction = this.protoDirection[direction];
        }
        this.getSprite().angle = this.d2a[direction];
    }
}

export default BaseElement;
