import { makeRequest, pprint } from 'utils/helpers';

class MenuState extends Phaser.State {

    create() {
        this.stage.backgroundColor = "#ffffff";
        this.game.canvas.style.border = "1px solid black";

        this.game.clearMenu();

        if (predefinedVars.currentUser.tanks.length == 0) {
            let h2 = document.createElement('h2');
            h2.textContent = "New one?";
            h2.style.textAlign = "center";
            this.game.menu.appendChild(h2);

            let h4 = document.createElement('h4');
            h4.textContent = "Create tank and go fight!";
            h4.style.textAlign = "center";
            this.game.menu.appendChild(h4);

            this.renderForm();
        } else {
            let h1 = document.createElement('h1');
            h1.textContent = "Select tank for fight";
            h1.style.textAlign = "center";
            this.game.menu.appendChild(h1);

            this.renderTanks();
        }
    }

    renderTanks() {
        this.loadTanks()
        .then(result => {
            result.every((el, i) => {
                if (i == 3) {
                    return false;
                }
                // TODO: Remove from here
                el.avatar = '/static/bbg_client/img/menu/tankist.jpg';
                this.createTankTab(this.game.menu, el);
                return true;
            });
        });
    }

    renderForm() {
        let callback = (e) => {
            e.preventDefault();
            let name = e.target.parentNode.querySelector('input').value;
            this.createTank(name)
            .then(result => {
                predefinedVars.currentUser.tanks = [result.tkey];
                this.game.state.restart();
            });
        };
        let html = `
            <div class="form-group">
                <label class="control-label" for="name">Tank name:</label>
                <input type="name" class="form-control" id="name" placeholder="Enter name">
                <span class="help-block"></span>
            </div>
            <button type="submit" class="btn btn-default">Submit</button>
        `;
        let blockEl = document.createElement('form');
        blockEl.innerHTML = html;

        blockEl.querySelector('button').addEventListener('click', callback);

        this.game.menu.appendChild(blockEl);
    }

    clearFormErrors() {
        let formGroup = document.querySelector('.form-group');
        formGroup.className = 'form-group';

        let errBlock = formGroup.querySelector('.help-block');
        errBlock.innerHTML = '';
    }

    renderFormErrors(errors) {
        let formGroup = document.querySelector('.form-group');
        formGroup.className = 'form-group has-error';

        let errBlock = formGroup.querySelector('.help-block');
        errBlock.innerHTML = '';

        let ulBlock = document.createElement('ul');
        errors.map((el) => {
            ulBlock.innerHTML += `<li>${el.detail}</li>`;
        });
        errBlock.appendChild(ulBlock);
        console.log(errors);
    }

    createTankTab(parent, data) {
        let callback = (e) => {
            e.preventDefault();
            this.game.state.start('GameState', true, true, data);
        };
        if (data.kda == null) {
            data.kda = '--';
        }
        let html = `
            <div style="border: 1px solid black; margin: 10px 0 0 0; height: 120px; border-radius: 10px;" data-tkey="${data.tkey}">
                <img src="${data.avatar}" width="76px" height="76px" style="float: left; margin: 24px 0 0 12px;">
                <span style="float: left; margin: 12px 0 12px 24px;"><b>Name: ${data.name}</b></span>
                <div style="float: left; margin: 0 0 0 24px;">
                    <div class="dropdown" style="float: left;">
                        <button class="dropbtn" style="background-color: grey;">Statistic</button>
                        <div class="dropdown-content">
                            <span><b>LVL: ${data.lvl}</b></span>
                            <span><b>Scores: ${data['scores-count']}</b></span>
                            <span><b>KDA: ${data.kda}</b></span>
                            <span><b>Kill count: ${data['kill-count']}</b></span>
                            <span><b>Death count: ${data['death-count']}</b></span>
                        </div>
                    </div>
                    <button class="btn btn-success" style="float: left; margin: 10px 0 0 24px;">SMASH!</button>
                </div>
            </div>
        `;
        let blockEl = document.createElement('div');
        blockEl.innerHTML = html;
        blockEl.className = 'row';

        parent.appendChild(blockEl);

        console.log(data);

        document.querySelector(`[data-tkey="${data.tkey}"] > div > button`).addEventListener('click', callback);
    }

    createTank(name) {
        this.clearFormErrors();
        return makeRequest({
            method: 'POST',
            type: 'tanks',
            url: 'tanks/',
            data: {
                name: name
            }
        })
        .then(result => {
            if (result.errors) {
                this.renderFormErrors(result.errors);
                return result;
            }
            return result.data.attributes;
        });
    }

    loadTanks() {
        return makeRequest({
            method: 'GET',
            type: 'tanks',
            url: 'user/tanks/'
        })
        .then(result => {
            return result.data.map((el) => el.attributes);
        });
    }

}

export default MenuState;
