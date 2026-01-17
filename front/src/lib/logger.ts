type LogLevel = 'info' | 'warn' | 'error' | 'debug';

class Logger {
    private format(level: LogLevel, message: string, data?: any) {
        const timestamp = new Date().toISOString();
        const logObj = {
            timestamp,
            level: level.toUpperCase(),
            message,
            ...(data && { data })
        };
        return JSON.stringify(logObj);
    }

    info(message: string, data?: any) {
        console.log(this.format('info', message, data));
    }

    warn(message: string, data?: any) {
        console.warn(this.format('warn', message, data));
    }

    error(message: string, error?: any) {
        // Handle Error objects specifically to extract stack and message
        const data = error instanceof Error ? {
            message: error.message,
            stack: error.stack,
            name: error.name
        } : error;

        console.error(this.format('error', message, data));
    }

    debug(message: string, data?: any) {
        if (process.env.NODE_ENV !== 'production') {
            console.debug(this.format('debug', message, data));
        }
    }
}

export const logger = new Logger();
