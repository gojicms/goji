# Goji Deployment

This directory contains the deployment configuration for Goji. The deployment process is designed to be simple and consistent across all environments.

## Deployment Process

1. Copy the deployment files:
   ```bash
   cp deploy/docker/docker-compose.yml.sample docker-compose.yml
   cp deploy/docker/Dockerfile.sample Dockerfile
   cp deploy/docker/deploy.sh.sample deploy.sh
   ```

2. Place your application code in the `application` directory:
   - Your custom templates
   - Your custom modules
   - Your configuration

3. Start the application:
   ```bash
   docker-compose up -d
   ```

4. Access the application at `http://localhost:8080`

## How It Works

1. **Dependencies**:
   - Core and contrib modules are fetched from GitHub during build
   - Your application code is the only local code needed

2. **Database**:
   - SQLite database for simplicity
   - Data persists in a Docker volume
   - Database container never goes down during deployments

3. **Application**:
   - Zero-downtime deployments
   - Health checks ensure stability
   - Automatic restarts on failure

## Production Deployment

For production deployments, you can optionally set up Git hooks:

```bash
cp deploy/docker/githooks/post-receive.sample .git/hooks/post-receive
chmod +x .git/hooks/post-receive
```

This will automatically deploy when you push to your repository.

## Customization

You can customize the deployment by modifying:

- `docker-compose.yml`: Adjust ports, environment variables, etc.
- `Dockerfile`: Change build process or base images
- `deploy.sh`: Modify deployment steps or add additional checks

## Notes

- All sample files have `.sample` extension - remove this when using them
- Make sure to set proper permissions on the deploy script and git hooks
- The application will automatically fetch the latest versions of core and contrib 