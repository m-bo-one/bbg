import { toFirstUpperCase } from 'utils/helpers';

class MainState extends Phaser.State {

    create() {
        this.stage.backgroundColor = "#ffffff";
        this.game.canvas.style.border = "1px solid black";

        this.game.clearMenu();

        let colEl = document.createElement('div');
        colEl.className = "col-md-offset-1 col-md-11";
        colEl.style.float = "none";
        colEl.style.marginTop = "100%";
        this.game.menu.append(colEl);

        this.socialButtonCreate(colEl, 'facebook');
        this.socialButtonCreate(colEl, 'github');
    }

    socialButtonCreate(parent, name) {
        let callback = () => {
            window.location.href = document.location.origin + predefinedVars.socialAuthURL[name];
        };
        let socName = toFirstUpperCase(name);
        let html = `<span class="fa fa-${name}"></span> Sign in with ${socName}`;
        let buttonEl = document.createElement('a');
        buttonEl.className = `btn btn-block btn-social btn-${name}`
        buttonEl.innerHTML = html;
        buttonEl.addEventListener('click', callback);
    
        parent.append(buttonEl);
    }

}

export default MainState;
