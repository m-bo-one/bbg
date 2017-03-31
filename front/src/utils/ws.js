import * as protobuf from "../../node_modules/protobufjs/index.js";

class ProtoStream {

    constructor(url, ...args) {
        this._ws = new WebSocket(url, ...args);
        this._ws.binaryType = "arraybuffer";
        this._ws.onopen = () => {
            console.log('open');
        };

        this._ws.onmessage = (e) => {
            if (e.data instanceof ArrayBuffer) {
                let bytearray = new Uint8Array(event.data);
                this.onmessage(bytearray);
            } else {
                throw new Error('Proto WS supports only protobuf message type.')
            }
        };
        this._ws.onclose = () => {
            console.log('restarting...');
            try {
                setTimeout(() => {
                    game.stream = new ProtoStream(url, ...args);
                    game.stream.proto = this.proto;
                }, 2000);
            } catch (e) {
                console.log('failed, trying again...');
            }
        };
        this.loadProtos({
            'protobufs/cmd.proto': {
                'enum': [
                    'Direction'
                ],
                'type': [
                    'BBGProtocol',
                    'TankUpdate',
                    'TankMove',
                ],
            }
        })
    }

    get connected() {
        return this._ws.readyState == this._ws.OPEN && this.proto.loaded;
    }

    get pbProtocol() {
        return this.proto.BBGProtocol;
    }

    send(type, data) {
        if (!this.connected) {
            setTimeout(() => {
                this.send(type, data);
            }, 1000);
            return;
        }
        let prData = {
            type: this.pbProtocol.Type[type] || this.pbProtocol.Type.UnhandledType,
            version: 1
        }
        if(data) {
            type = ((string) => string.charAt(0).toLowerCase() + string.slice(1))(type);
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
            console.log('decoded: ', decoded);
            game.state.states[game.state.current].wsUpdate(decoded);
        }
    }

    loadProtos(pdata) {
        this.proto = {
            loaded: false
        };
        let _length = Object.keys(pdata).length;
        Object.keys(pdata).forEach(pkey => {
            protobuf.load(pkey, (err, root) => {
                _length--;
                if (err) {
                    console.log("Error during protobuf loading. ", err);
                    return;
                }
                let tdata = pdata[pkey];

                Object.keys(tdata).forEach(tkey => {
                    let protoNames = tdata[tkey];

                    protoNames.forEach(protoName => {
                        try {
                            switch (tkey) {
                                case 'enum':
                                    this.proto[protoName] = root.lookupEnum(
                                        `proto.${protoName}`
                                    );
                                    return;
                                case 'type':
                                    this.proto[protoName] = root.lookupType(
                                        `proto.${protoName}`
                                    );
                                    return;
                            }
                        } catch(e) {
                            console.log(`Error during setup. Detailed: ${e} - ${protoName}.`);
                        }
                    });
                });
                if (_length === 0) {
                    this.proto.loaded = true;
                }
            });
        });
    }

    onLoadComplete(func) {
        if (!this.connected) {
            setTimeout(() => {
                this.onLoadComplete(func);
            }, 1000);
            return;
        }
        func();
    }

}

export default ProtoStream;