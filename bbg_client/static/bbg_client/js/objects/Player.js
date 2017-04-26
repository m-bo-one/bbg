class Player {

    constructor(game) {
        this.game = game;
        this._elements = {};
    }

    get tank() {
        return this._elements.tank;
    }

    get(name) {
        return this._elements[name];
    }

    add(name, element) {
        this._elements[name] = element;
    }

    remove(name) {
        delete this._elements[name];
    }

    clone(name, obj) {
        return Object.assign(new Player(), obj);
    }

    update(data) {
        for (let k in this._elements) {
            this._elements[k].update(data);
        }
    }

}

export default Player;
