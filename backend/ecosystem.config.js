module.exports = {
    apps: [
        {
            name: "whatsapp-bot",
            script: "./bot-linux-amd64",
            instances: 1,
            autorestart: true,
            watch: false,
            max_memory_restart: "800M", // 1GB limit safety
            env: {
                NODE_ENV: "production",
                // Inherits system vars, but can override here
            }
        }
    ]
};
