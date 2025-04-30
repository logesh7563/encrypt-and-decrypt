const net = require('net');
const id = 'img1';
const client = net.createConnection(8084, 'localhost', () => {
  // build header
  const idBuf = Buffer.from(id, 'utf8');
  const header = Buffer.alloc(1 + 4);
  header.writeUInt8(1, 0);                  // ImageDataRequest
  header.writeUInt32BE(idBuf.length, 1);    // length
  client.write(Buffer.concat([header, idBuf]));
});

let state = 0, len = 0, chunks = [];
client.on('data', chunk => {
  chunks.push(chunk);
  const buf = Buffer.concat(chunks);

  if (state === 0 && buf.length >= 1) {
    if (buf.readUInt8(0) !== 2) {
      console.error('Unexpected msg type', buf.readUInt8(0));
      client.end();
      return;
    }
    state = 1;
  }
  if (state === 1 && buf.length >= 5) {
    len = buf.readUInt32BE(1);
    state = 2;
  }
  if (state === 2 && buf.length >= 5 + len) {
    const data = buf.slice(5, 5 + len);
    console.log(`Received ${data.length} bytes of encrypted image.`);
    // e.g. write to disk:
    require('fs').writeFileSync('img1.enc', data);
    client.end();
  }
});
client.on('end', () => console.log('Done.'));