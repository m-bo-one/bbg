import BaseElement from 'objects/BaseElement';
import HealthBar from 'objects/HealthBar';
import ScoreBoard from 'objects/ScoreBoard';

class HUD extends BaseElement {

    constructor(game, data, group) {
        super(game, data);
        this.tankId = data.tankId;
        this.group = group;
        this.createHealthBar(data);
        this.createScoreBoard(data);
        // this.createStatBlock(data);
    }

    // createOrUpdatePing(ping=0) {
    //     if (!this._pingText) {
    //         this._pingText = this.game.add.text(0, 20, `Ping: ${ping}`, this.frontLayer);
    //         this._pingText.x = this.game.width - this._pingText.width - 20;
    //         this._pingText.fixedToCamera = true; 
    //     } else {
    //         this._pingText.text = `Ping: ${ping}`;
    //     }
    // }

    get tank() {
        return this.game.currentTank;
    }

    isPlayerTank() {
        return this.tank.id === this.tankId;
    }

    update(data) {
        if (!this.isPlayerTank()) return;

        if (data.health != this.tank.health) {
            if (data.health < 0) {
                this._healthBar.setWidth(0);
            } else {
                this._healthBar.setWidth(data.health * 2);
            }
        }
        if (this.tank.isDead()) {
            this.tank.changeColor(0xff9a22);
        } else {
            // HARDCODED VALUE OF COLOR
            this.tank.changeColor(16777215);
        }
    }

    destroy() {
        if (!this.isPlayerTank()) return;

        this._healthBar.kill();
        this._killStatBlock();
    }

    createHealthBar(data) {
        this._healthBar = new HealthBar(this.game, {
            width: 100 * 2,
            height: 20,
            bar: {
              color: 'red'
            },
            animationDuration: 500,
        });
        this._healthBar.setPosition(130, this.game.height - 50);
        this._healthBar.setWidth(data.health * 2);
        this._healthBar.setFixedToCamera(true);

        this.group.add(this._healthBar.bgSprite);
        this.group.add(this._healthBar.barSprite);
    }

    createScoreBoard(data) {
        this._scoreBoard = new ScoreBoard(this.game);
        this._scoreBoard.setGroup(this.group);
    }

    // createStatBlock(data) {
    //     let initX = 30;
    //     let initY = 20;
    //     let offset = 30;
    //     this._scoreText = this.game.add.text(initX, initY, `Scores: ${data["scores-count"]}`, this.group);
    //     this._killText = this.game.add.text(initX, initY + offset, `Kills: ${data["kill-count"]}`, this.group);
    //     this._deathText = this.game.add.text(initX, initY + 2 * offset, `Death: ${data["death-count"]}`, this.group);

    //     this._scoreText.fixedToCamera = true; 
    //     this._killText.fixedToCamera = true; 
    //     this._deathText.fixedToCamera = true; 
    // }

    // killStatBlock() {
    //     this._scoreText.destroy(); 
    //     this._killText.destroy(); 
    //     this._deathText.destroy();        
    // }

    createRespawnBlock(data) {
        if (typeof this._restartText === "object") return;
        data.counter = data.counter || 3;
        this._restartText = this.game.add.text(0, 0, `Respawn at: ${data.counter}`, this.group);
        this._restartText.x = this.game.width - 50 - this._restartText.width;
        this._restartText.y = this.game.height - 75;
        this._restartText.fixedToCamera = true; 
        let id = setInterval(() => {
            counter--;
            if (counter === 0) {
                clearInterval(id);
                this._restartText.destroy();
                delete this._restartText;
                return
            }
            this._restartText.text = `Respawn at: ${counter}`;
        }, 1000);
    }

}

export default HUD;
