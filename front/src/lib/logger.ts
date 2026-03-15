type LogLevel = 'info' | 'warn' | 'error' | 'debug';

class Logger {
    private format(level: LogLevel, message: string, data?: unknown) {
        const timestamp = new Date().toISOString();
        const logObj = {
            timestamp,
            level: level.toUpperCase(),
            message,
            ...(typeof data === 'object' && data !== null ? data : { data })
        };
        return JSON.stringify(logObj);
    }

    info(message: string, data?: unknown) {
        console.log(this.format('info', message, data));
    }

    warn(message: string, data?: unknown) {
        console.warn(this.format('warn', message, data));
    }

    error(message: string, error?: unknown) {
        // Handle Error objects specifically to extract stack and message
        const data = error instanceof Error ? {
            message: error.message,
            stack: error.stack,
            name: error.name
        } : error;

        console.error(this.format('error', message, data));
    }

    debug(message: string, data?: unknown) {
        if (process.env.NODE_ENV !== 'production') {
            console.debug(this.format('debug', message, data));
        }
    }
}

export const logger = new Logger();
