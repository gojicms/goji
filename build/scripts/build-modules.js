import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { mkdir, readdir } from 'fs/promises';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, '../..');
const modulesDir = join(rootDir, 'dist/modules');
const contribDir = join(rootDir, 'contrib');

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

  return new Promise((resolve, reject) => {
    build.on('error', reject);
    build.on('close', (code) => {
      if (code === 0) resolve();
      else reject(new Error(`Build failed with code ${code}`));
    });
  });
}

async function main() {
  // Ensure modules directory exists
  await mkdir(modulesDir, { recursive: true });

  // Get all directories in contrib
  const entries = await readdir(contribDir, { withFileTypes: true });
  const modules = entries
    .filter(entry => entry.isDirectory())
    .map(dir => join(contribDir, dir.name));

  // Build each module
  for (const module of modules) {
    try {
      await buildModule(module);
    } catch (err) {
      console.error(`Failed to build module ${module}:`, err);
    }
  }
}

main().catch(console.error); 