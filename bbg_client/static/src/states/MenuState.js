import { makeRequest } from 'utils/helpers';

class MenuState extends Phaser.State {

    preload() {
        this.game.imageLoad('new', 'assets/menu/new.png');
        this.game.imageLoad('tankist', 'assets/menu/tankist.jpg');
    }

    create() {
        this.stage.backgroundColor = "#ffffff";
        this.game.canvas.style.border = "1px solid black";

        this.menuGroup = this.game.add.group();

        let header = this.game.add.text(this.game.world.centerX, 80, 'SELECT TANK');
        header.anchor.set(0.5);
        this.menuGroup.add(header);

        let button;

        if (!predefinedVars.currentUser.tanks) {
            button = this.hardcodedNewButton(this.createTank);
        } else {
            button = this.hardcodedTankBorder();
        }
        this.menuGroup.add(button);
    }

    createTank() {
        makeRequest('tanks', 'POST', {'name': null})
        .then(function(result) {
            console.log(result);
        });
    }

    hardcodedNewButton(callback) {
        let offset = this.game.world.centerY;
        let bGr = this.game.add.group();

        let diff = 50;

        let newButton = this.game.add.button(this.game.world.centerX - 50 - diff, offset, 'new', callback, this, 2, 1, 0);
        newButton.scale.setTo(0.15);
        newButton.anchor.setTo(0.5);

        bGr.add(newButton);

        let text = this.game.add.text(this.game.world.centerX + 100 - diff, offset, 'Create new tank');
        text.anchor.setTo(0.5);

        bGr.add(text);

        let border = game.add.graphics(this.game.world.centerX - 100 - diff, offset - 50);
        border.lineStyle(1, "black", 1);
        border.drawRect(0, 0, 350, 100);

        bGr.add(border);

        bGr.scale.setTo(0.75);
        bGr.x += 110;

        return bGr
    }

    hardcodedTankBorder(tankName='HERMAN', tankLvl=1) {
        let offset = this.game.world.centerY;
        let bGr = this.game.add.group();

        let diff = 50;

        let imgTankist = this.game.add.sprite(this.game.world.centerX - diff - 50, offset, 'tankist');
        imgTankist.scale.setTo(0.15);
        imgTankist.anchor.setTo(0.5);

        bGr.add(imgTankist);

        let textName = this.game.add.text(this.game.world.centerX + 100 - diff, offset - 20, tankName);
        textName.anchor.setTo(0.5);

        bGr.add(textName);

        let textLvl = this.game.add.text(this.game.world.centerX + 100 - diff, offset + 20, `LVL: ${tankLvl}`);
        textLvl.anchor.setTo(0.5);

        bGr.add(textLvl);

        let border = game.add.graphics(this.game.world.centerX - 100 - diff, offset - 50);
        border.lineStyle(1, "black", 1);
        border.drawRect(0, 0, 350, 100);

        bGr.add(border);

        bGr.scale.setTo(0.75);
        bGr.x += 110;

        return bGr
    }

}

export default MenuState;
