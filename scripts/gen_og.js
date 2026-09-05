#!/usr/bin/env node
// Generate a 1200x630 OG PNG with duo gradient + simple shapes (no text).
// Output: website/public/og.png
const fs = require("fs");
const zlib = require("zlib");

const W = 1200, H = 630;
const buf = Buffer.alloc(W * H * 3); // RGB

function setPx(x, y, r, g, b) {
  if (x < 0 || x >= W || y < 0 || y >= H) return;
  const i = (y * W + x) * 3;
  buf[i] = r; buf[i + 1] = g; buf[i + 2] = b;
}

function lerp(a, b, t) { return Math.round(a + (b - a) * t); }

// background dark gradient top-left -> bottom-right
for (let y = 0; y < H; y++) {
  for (let x = 0; x < W; x++) {
    const tx = x / W, ty = y / H;
    let r = lerp(5, 12, tx);
    let g = lerp(5, 14, ty);
    let b = lerp(8, 24, (tx + ty) / 2);
    // radial cyan glow top-left
    const d1 = Math.hypot(x - 360, y - 130);
    const k1 = Math.max(0, 1 - d1 / 600);
    r = lerp(r, 0, k1 * 0.35 * 0.5);
    g = lerp(g, 210, k1 * 0.35);
    b = lerp(b, 255, k1 * 0.35);
    // radial magenta glow bottom-right
    const d2 = Math.hypot(x - 1020, y - 530);
    const k2 = Math.max(0, 1 - d2 / 480);
    r = lerp(r, 255, k2 * 0.30);
    g = lerp(g, 42, k2 * 0.30 * 0.4);
    b = lerp(g, 133, k2 * 0.30 * 0.6);
    setPx(x, y, r, g, b);
  }
}

// draw cyan->magenta bar (logo mark) top-left at (80, 100), 72x72 with rounded corners (approx)
function fillRect(x0, y0, w, h, r, g, b) {
  for (let y = y0; y < y0 + h; y++) for (let x = x0; x < x0 + w; x++) setPx(x, y, r, g, b);
}
function fillRound(x0, y0, w, h, r, g, b, rad = 14) {
  for (let y = y0; y < y0 + h; y++) {
    for (let x = x0; x < x0 + w; x++) {
      let dx = 0, dy = 0;
      if (x < x0 + rad) dx = x0 + rad - x;
      else if (x >= x0 + w - rad) dx = x - (x0 + w - rad - 1);
      if (y < y0 + rad) dy = y0 + rad - y;
      else if (y >= y0 + h - rad) dy = y - (y0 + h - rad - 1);
      if (dx * dx + dy * dy > rad * rad) continue;
      setPx(x, y, r, g, b);
    }
  }
}
// logo gradient block 72x72 rounded
const logoSize = 72, logoX = 80, logoY = 100;
for (let y = logoY; y < logoY + logoSize; y++) {
  for (let x = logoX; x < logoX + logoSize; x++) {
    let dx = 0, dy = 0;
    if (x < logoX + 14) dx = logoX + 14 - x;
    else if (x >= logoX + logoSize - 14) dx = x - (logoX + logoSize - 15);
    if (y < logoY + 14) dy = logoY + 14 - y;
    else if (y >= logoY + logoSize - 14) dy = y - (logoY + logoSize - 15);
    if (dx * dx + dy * dy > 14 * 14) continue;
    const t = (x - logoX) / logoSize;
    const cr = lerp(0, 255, t);
    const cg = lerp(210, 42, t);
    const cb = lerp(255, 133, t);
    setPx(x, y, cr, cg, cb);
  }
}

// horizontal duo gradient bar at bottom (visual separator)
const barY = 480, barH = 4;
for (let x = 80; x < 1120; x++) {
  const t = (x - 80) / 1040;
  const cr = lerp(0, 255, t);
  const cg = lerp(210, 42, t);
  const cb = lerp(255, 133, t);
  for (let y = barY; y < barY + barH; y++) setPx(x, y, cr, cg, cb);
}

// emit PNG
function chunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length, 0);
  const typeBuf = Buffer.from(type, "ascii");
  const crc = Buffer.alloc(4);
  // simple CRC32
  let c = 0xffffffff;
  for (const b of Buffer.concat([typeBuf, data])) {
    c ^= b;
    for (let k = 0; k < 8; k++) c = (c & 1) ? (c >>> 1) ^ 0xedb88320 : (c >>> 1);
  }
  crc.writeUInt32BE((c ^ 0xffffffff) >>> 0, 0);
  return Buffer.concat([len, typeBuf, data, crc]);
}

const sig = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
const ihdr = Buffer.alloc(13);
ihdr.writeUInt32BE(W, 0);
ihdr.writeUInt32BE(H, 4);
ihdr[8] = 8;       // bit depth
ihdr[9] = 2;       // color type RGB
ihdr[10] = 0;      // compression
ihdr[11] = 0;      // filter
ihdr[12] = 0;      // interlace

// scanlines with filter byte 0
const raw = Buffer.alloc(H * (1 + W * 3));
for (let y = 0; y < H; y++) {
  raw[y * (1 + W * 3)] = 0;
  buf.copy(raw, y * (1 + W * 3) + 1, y * W * 3, (y + 1) * W * 3);
}
const idat = zlib.deflateSync(raw);

const png = Buffer.concat([
  sig,
  chunk("IHDR", ihdr),
  chunk("IDAT", idat),
  chunk("IEND", Buffer.alloc(0)),
]);
fs.writeFileSync("E:/javaproject/my/2026/agent_team/website/public/og.png", png);
console.log("og.png:", png.length, "bytes");