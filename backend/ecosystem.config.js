module.exports = {
  apps: [
    {
      name: 'family-finance-backend',
      script: './bin/api',
      cwd: __dirname,
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      max_restarts: 10,
      restart_delay: 3000,
      kill_timeout: 10000,
      env: {
        NODE_ENV: 'production',
        APP_ENV: 'production',
      },
      out_file: './logs/pm2/backend.out.log',
      error_file: './logs/pm2/backend.err.log',
      merge_logs: true,
      time: true,
    },
  ],
}
