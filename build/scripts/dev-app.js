import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { mkdir, cp } from 'fs/promises';
import chokidar from 'chokidar';

const __dirname = dirname(fileURLToPath(import.meta.url));
const rootDir = join(__dirname, '../..');
const distDir = join(rootDir, 'dist.dev');
const appBinary = join(distDir, 'goji');

async function copyStaticFiles() {
  // Copy admin and web folders
  await Promise.all([
    cp(join(rootDir, 'application/admin'), join(distDir, 'admin'), { recursive: true }),
    cp(join(rootDir, 'application/web'), join(distDir, 'web'), { recursive: true })
  ]);
}

async function buildApp() {
  console.log('Building application...');
  const build = spawn('go', [
    'build',
    '-o',
    appBinary,
    'application/main.go'
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
  // Create dist.dev directory if it doesn't exist
  await mkdir(distDir, { recursive: true });
  
  // Copy static files
  await copyStaticFiles();

  // Initial build
  await buildApp();

  // Start the application
  let app = spawn(appBinary, [], {
    cwd: distDir,
    stdio: 'inherit',
    env: {
      ...process.env,
      GO_ENV: 'development',
      DIST_DIR: distDir
    }
  });

  app.on('error', (err) => {
    console.error('Failed to start application:', err);
    process.exit(1);
  });

  // Watch for changes in application, core, and modules directories
  const watcher = chokidar.watch([
    join(rootDir, 'application/**/*.go'),
    join(rootDir, 'core/**/*.go'),
    join(rootDir, 'application/admin/**/*'),
    join(rootDir, 'application/web/**/*'),
    join(distDir, 'modules/**/*.so')  // Watch for module changes
  ], {
    ignored: /(^|[\/\\])\../,
    persistent: true
  });

  watcher.on('change', async (path) => {
    console.log(`File ${path} has been changed, rebuilding application...`);
    
    if (path.includes('admin/') || path.includes('web/')) {
      // If it's a static file change, copy it
      const relativePath = path.replace(join(rootDir, 'application/'), '');
      const destPath = join(distDir, relativePath);
      await mkdir(dirname(destPath), { recursive: true });
      await cp(path, destPath);
      console.log(`Copied ${path} to ${destPath}`);
    } else {
      // Rebuild and restart the application
      try {
        await buildApp();
        app.kill();
        const newApp = spawn(appBinary, [], {
          cwd: distDir,
          stdio: 'inherit',
          env: {
            ...process.env,
            GO_ENV: 'development',
            DIST_DIR: distDir
          }
        });
        newApp.on('error', (err) => {
          console.error('Failed to restart application:', err);
        });
        app = newApp;
      } catch (err) {
        console.error('Failed to rebuild application:', err);
      }
    }
  });

  process.on('SIGINT', () => {
    app.kill();
    process.exit();
  });
}

main().catch(console.error); 