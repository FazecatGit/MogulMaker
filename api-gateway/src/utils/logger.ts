import fs from 'fs';
import path from 'path';

class Logger {
  private logDir: string;

  constructor() {
    this.logDir = path.join(process.cwd(), 'logs');
    if (!fs.existsSync(this.logDir)) {
      fs.mkdirSync(this.logDir, { recursive: true });
    }
  }

  private getTimestamp(): string {
    return new Date().toISOString();
  }

  private formatMessage(level: string, message: string, data?: any): string {
    const timestamp = this.getTimestamp();
    const dataStr = data ? ` ${JSON.stringify(data)}` : '';
    return `[${timestamp}] [${level}] ${message}${dataStr}`;
  }

  info(message: string, data?: any): void {
    const msg = this.formatMessage('INFO', message, data);
    console.log(msg);
    this.writeToFile(msg);
  }

  error(message: string, data?: any): void {
    const msg = this.formatMessage('ERROR', message, data);
    console.error(msg);
    this.writeToFile(msg);
  }

  warn(message: string, data?: any): void {
    const msg = this.formatMessage('WARN', message, data);
    console.warn(msg);
    this.writeToFile(msg);
  }

  debug(message: string, data?: any): void {
    if (process.env.DEBUG === 'true') {
      const msg = this.formatMessage('DEBUG', message, data);
      console.log(msg);
      this.writeToFile(msg);
    }
  }

  private writeToFile(message: string): void {
    const logFile = path.join(this.logDir, `app-${new Date().toISOString().split('T')[0]}.log`);
    fs.appendFileSync(logFile, message + '\n');
  }
}

export default new Logger();
