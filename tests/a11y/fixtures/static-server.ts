// fixtures/static-server.ts - serves ./website/out on an ephemeral port.
//
// Used by Playwright + axe-core to crawl the static export without
// spinning up a real Next.js server.

import { createServer, type Server } from 'node:http';
import { readFile, stat } from 'node:fs/promises';
import { extname, join, normalize } from 'node:path';

export type StaticServer = {
  url: string;
  close: () => Promise<void>;
};

const MIME: Record<string, string> = {
  '.html': 'text/html; charset=utf-8',
  '.js': 'application/javascript; charset=utf-8',
  '.css': 'text/css; charset=utf-8',
  '.json': 'application/json; charset=utf-8',
  '.svg': 'image/svg+xml',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.jpeg': 'image/jpeg',
  '.webp': 'image/webp',
  '.ico': 'image/x-icon',
  '.woff': 'font/woff',
  '.woff2': 'font/woff2',
  '.txt': 'text/plain; charset=utf-8',
  '.xml': 'application/xml; charset=utf-8',
};

async function resolveFile(rootDir: string, urlPath: string): Promise<string | null> {
  // Map "/" -> "/index.html"; "/en/" -> "/en/index.html"; etc.
  const cleaned = urlPath.split('?')[0].split('#')[0];
  const candidates = [cleaned, join(cleaned, 'index.html')];
  for (const c of candidates) {
    const abs = normalize(join(rootDir, c));
    if (!abs.startsWith(rootDir)) continue;
    try {
      const s = await stat(abs);
      if (s.isFile()) return abs;
    } catch {
      // try next
    }
  }
  return null;
}

export async function startStaticServer(): Promise<StaticServer> {
  const rootDir = normalize(join(process.cwd(), 'website/out'));
  const server: Server = createServer(async (req, res) => {
    try {
      const urlPath = req.url ?? '/';
      const file = await resolveFile(rootDir, urlPath);
      if (!file) {
        res.statusCode = 404;
        res.end('not found');
        return;
      }
      const buf = await readFile(file);
      res.setHeader('Content-Type', MIME[extname(file).toLowerCase()] ?? 'application/octet-stream');
      res.end(buf);
    } catch (e) {
      res.statusCode = 500;
      res.end(String(e));
    }
  });

  await new Promise<void>((resolve) => server.listen(0, '127.0.0.1', resolve));
  const addr = server.address();
  if (!addr || typeof addr === 'string') throw new Error('static-server: bad address');

  return {
    url: `http://127.0.0.1:${addr.port}`,
    close: () =>
      new Promise<void>((resolve, reject) =>
        server.close((err) => (err ? reject(err) : resolve())),
      ),
  };
}
