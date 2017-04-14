class MainState extends Phaser.State {

    preload() {
        this.game.imageLoad('github', 'assets/social/github.png');
        this.game.imageLoad('facebook', 'assets/social/facebook.png');
    }

    create() {
        this.stage.backgroundColor = "#ffffff";
        this.game.canvas.style.border = "1px solid black";
        this.buttonGroup = this.game.add.group();

        let offset = this.game.world.centerY;

        let header = this.game.add.text(this.game.world.centerX, offset - 40, 'BBG TANKS');
        header.anchor.set(0.5);
        this.buttonGroup.add(header);

        this.socialButtonCreate(this.game.world.centerX, offset, 'github');
        this.socialButtonCreate(this.game.world.centerX, offset + 40, 'facebook');
    }

    update() {

    }

    render() {

    }

    socialButtonCreate(x, y, name, scale=0.5) {
        let callback = () => {
            window.location.href = predefinedVars.socialAuthURL[name];
        };
        let icon = this.game.add.button(x, y, name, callback, this, 2, 1, 0);
        icon.scale.setTo(scale);
        icon.anchor.setTo(scale);
        this.buttonGroup.add(icon);
    }

}

export default MainState;
