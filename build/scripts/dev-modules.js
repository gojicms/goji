import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { mkdir } from 'fs/promises';
import chokidar from 'chokidar';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, '../..');
const modulesDir = join(rootDir, 'dist.dev/modules');

async function buildModule(modulePath) {
  const moduleName = modulePath.split('/').pop();
  console.log(`Building module: ${moduleName}`);
  
  const build = spawn('go', [
    'build',
    '-buildmode=plugin',
    '-o',
    join(modulesDir, `${moduleName}.so`),
    modulePath
  ], {
    cwd: rootDir,
    stdio: 'inherit'
  });

  build.on('error', (err) => {
    console.error(`Failed to build module ${moduleName}:`, err);
  });
}

async function main() {
  // Ensure modules directory exists
  await mkdir(modulesDir, { recursive: true });

  // Watch for changes in modules
  const watcher = chokidar.watch(join(rootDir, 'contrib/documents/**/*.go'), {
    ignored: /(^|[\/\\])\../,
    persistent: true
  });

  watcher.on('change', (path) => {
    console.log(`File ${path} has been changed`);
    buildModule(join(rootDir, 'contrib/documents'));
  });

  // Initial build
  buildModule(join(rootDir, 'contrib/documents'));
}

main().catch(console.error); 