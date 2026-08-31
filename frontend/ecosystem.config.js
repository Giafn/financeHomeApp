module.exports = {
  apps: [
    {
      name: 'family-finance-frontend',
      script: './node_modules/next/dist/bin/next',
      args: 'start -p 3000',
      cwd: __dirname,
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      max_restarts: 10,
      restart_delay: 3000,
      kill_timeout: 10000,
      env: {
        NODE_ENV: 'production',
      },
      out_file: './logs/pm2/frontend.out.log',
      error_file: './logs/pm2/frontend.err.log',
      merge_logs: true,
      time: true,
    },
  ],
}
