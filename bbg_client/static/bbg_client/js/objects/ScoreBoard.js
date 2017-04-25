/**
 Copyright (c) 2017 Bogdan Kurinnyi (bogdankurinniy.dev1@gmail.com)

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
 */

var ScoreBoard = function(game) {
    this.game = game;
    this.group = game.add.group();
    this.group.inputEnableChildren = true;

    var data = {
        R: 10,
        x: game.width - 210,
        y: 10,
        margin: {
            x: 0,
            y: 30
        },
        width: 200,
        height: 20,
        color: {
            border: 0x545454,
            rect: 0x8A2BE2
        }
    };

    this._offset = {
        x: 0,
        y: 0
    };
    this.group.onChildInputOver.add(item => {
        item.graphicsData[0].fillAlpha = 0.8;
    }, this.game);
    this.group.onChildInputOut.add(item => {
        item.graphicsData[0].fillAlpha = 1;
    }, this.game);

    this._drawBox(data);
    this._drawBox(data);
    this._drawBox(data);
    this._drawBox(data);

    this.setFixedToCamera(true);
};
ScoreBoard.prototype.constructor = ScoreBoard;

ScoreBoard.prototype._drawBox = function(data) {
    var gr = this.game.add.graphics();
    gr.beginFill(data.color.rect);
    gr.lineStyle(3, data.color.border, 1);
    gr.drawRoundedRect(data.x + this._offset.x, data.y + this._offset.y, data.width, data.height, data.R);
    gr.endFill();

    this._offset.x += data.margin.x;
    this._offset.y += data.margin.y;

    this.group.add(gr);
};

ScoreBoard.prototype.setGroup = function(group) {
    group.add(this.group);
};

ScoreBoard.prototype.setPosition = function(x, y) {

};

ScoreBoard.prototype.setFixedToCamera = function(fixedToCamera) {
    this.group.fixedToCamera = fixedToCamera;
};

ScoreBoard.prototype.kill = function() {

};

module.exports = ScoreBoard;
