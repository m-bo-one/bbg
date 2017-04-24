import * as protobuf from "../../../../node_modules/protobufjs/index.js";
import { pprint, toFirstLowerCase } from 'utils/helpers';

class ProtoStream {

    constructor(url, callback=null, ...args) {
        game.flushMap();

        this._url = url;
        this._callback = callback;

        this._ws = new WebSocket(url, ...args);
        this._ws.binaryType = "arraybuffer";

        this._ws.onopen = () => {
            pprint('open');
        };

        this._ws.onmessage = (e) => {
            if (e.data instanceof ArrayBuffer) {
                let bytearray = new Uint8Array(e.data);
                this.onmessage(bytearray);
            } else {
                throw new Error('Proto WS supports only protobuf message type.')
            }
        };
        this._ws.onclose = (e) => {
            pprint('ws closed', e);
            this.retry(url, callback, ...args);
        };
        this._ws.onerror = (e) => {
            pprint('ws error', e);
        };
        this.loadProtos({
            'bbg1': {
                'enum': [
                    'Direction'
                ],
                'type': [
                    'BBGProtocol',
                    'TankUpdate',
                    'TankMove',
                    'TankShoot',
                    'BulletUpdate',
                    'Heartbeat',
                ],
            }
        })
        this.onLoadComplete();
    }

    retry(url, callback, ...args) {
        game.stream._ws.close();
        try {
            game.currentState.stopHeartbeat();
        } catch(e) {
            console.log(e);
        }
        console.log('restarting...');
        try {
            setTimeout(() => {
                game.stream = new ProtoStream(url, callback, ...args);
                pprint('successfully restarted.');
            }, 2000);
        } catch (e) {
            pprint('failed, trying again...');
        }
    }

    get connected() {
        return this._ws.readyState == this._ws.OPEN && this.proto.loaded;
    }

    get pbProtocol() {
        return this.proto.BBGProtocol;
    }

    get proto() {
        return this._proto.bbg1;
    }

    send(type, data) {
        if (!this.connected) {
            setTimeout(() => {
                this.send(type, data);
            }, 1000);
            return;
        }
        let prData = {
            type: this.pbProtocol.Type[`T${type}`] || this.pbProtocol.Type.UnhandledType,
            version: 1
        }
        if(typeof data !== undefined) {
            type = toFirstLowerCase(type);
            prData[type] = data;
        }
        // console.log("Obj2pb: ", prData);
        let msg = this.pbProtocol.fromObject(prData);
        let encoded = this.pbProtocol.encode(msg).finish();
        this._ws.send(encoded);
    }

    onmessage(bytearray) {
        let decoded = this.pbProtocol.decode(bytearray);
        if (Object.keys(decoded).length != 0) {
            // pprint('decoded: ', decoded);
            game.currentState.wsUpdate(decoded);
        }
    }

    getPing() {
        this.send('Ping', {timestamp: Math.floor(Date.now() / 1000)});
    }

    loadProtos(pdata) {
        this._proto = {};
        let _length = Object.keys(pdata).length;
        Object.keys(pdata).forEach(pkey => {
            protobuf.load(`static/protobufs/${pkey}.proto`, (err, root) => {
                _length--;
                if (err) {
                    pprint("Error during protobuf loading. ", err);
                    return;
                }
                let tdata = pdata[pkey];

                this._proto[pkey] = {
                    loaded: false
                };

                Object.keys(tdata).forEach(tkey => {
                    let protoNames = tdata[tkey];

                    protoNames.forEach(protoName => {
                        try {
                            switch (tkey) {
                                case 'enum':
                                    this._proto[pkey][protoName] = root.lookupEnum(
                                        `${pkey}.${protoName}`
                                    );
                                    return;
                                case 'type':
                                    this._proto[pkey][protoName] = root.lookupType(
                                        `${pkey}.${protoName}`
                                    );
                                    return;
                            }
                        } catch(e) {
                            pprint(`Error during setup. Detailed: ${e} - ${protoName}.`);
                        }
                    });
                });
                if (_length === 0) {
                    this._proto[pkey].loaded = true;
                }
            });
        });
    }

    onLoadComplete() {
        if (!this.connected) {
            setTimeout(() => {
                this.onLoadComplete();
            }, 1000);
            return;
        }
        pprint('Stream load completed.')
        if (this._callback !== null) this._callback();
    }

}

export default ProtoStream;