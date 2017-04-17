import { makeRequest } from 'utils/helpers';

class MenuState extends Phaser.State {

    create() {
        this.stage.backgroundColor = "#ffffff";
        this.game.canvas.style.border = "1px solid black";

        this.game.clearMenu();

        if (!predefinedVars.currentUser.tanks) {
            let h1 = document.createElement('h1');
            h1.textContent = "SELECT TANK";
            this.game.menu.append(h1);
        } else {
            this.retrieveTanks();
            this.createTankTab(this.game.menu, {lvl: 1, tkey: 232, avatar: '/static/bbg_client/img/menu/tankist.jpg'});
        }
    }

    createTankTab(parent, data) {
        let html = `
            <div style="border: 1px solid black; margin-top: 48px;" data-tkey="${data.tkey}">
                <img src="${data.avatar}" width="76px" height="76px">
                <span style="margin-left: 48px;"><b>LVL: ${data.lvl}</b></span>
                <button class="btn btn-success" style="margin-left: 48px;">Connect</button>
            </div>
        `;
        let blockEl = document.createElement('div');
        blockEl.innerHTML = html;
        blockEl.className = 'row';

        parent.append(blockEl);
    }

    createTank() {
        makeRequest({
            method: 'POST',
            type: 'tanks',
            url: 'tanks/',
            data: {
                name: 'test'
            }
        })
        .then(function(result) {
            console.log(result);
        })
        .catch(function(error) {
            console.log('Request failed', error);
        });
    }

    retrieveTanks() {
        makeRequest({
            method: 'GET',
            type: 'tanks',
            url: 'user/tanks/'
        })
        .then(function(result) {
            console.log(result);
        })
        .catch(function(error) {
            console.log('Request failed', error);
        });
    }

}

export default MenuState;
