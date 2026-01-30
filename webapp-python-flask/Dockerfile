# syntax=docker/dockerfile:1.4

# Production stage with Nginx and Python
FROM nginx:alpine

# Install Python, pip, and supervisor to run both nginx and Flask
RUN apk add --no-cache python3 py3-pip supervisor

# Set up Python virtual environment to avoid conflicts
RUN python3 -m venv /opt/venv
ENV PATH="/opt/venv/bin:$PATH"

# Copy and install Python dependencies
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY app.py .
COPY templates/ templates/
COPY static/ static/

# Copy nginx configuration
COPY .nginx/nginx.conf /etc/nginx/conf.d/default.conf

# Copy static files to nginx directory
COPY static/ /usr/share/nginx/html/static/

# Create supervisor configuration
RUN mkdir -p /etc/supervisor/conf.d
RUN echo $'[supervisord]\n\
nodaemon=true\n\
user=root\n\
\n\
[program:nginx]\n\
command=nginx -g "daemon off;"\n\
autostart=true\n\
autorestart=true\n\
stdout_logfile=/dev/stdout\n\
stdout_logfile_maxbytes=0\n\
stderr_logfile=/dev/stderr\n\
stderr_logfile_maxbytes=0\n\
\n\
[program:flask]\n\
command=/opt/venv/bin/gunicorn --bind 0.0.0.0:5000 --workers 2 app:app\n\
directory=/app\n\
autostart=true\n\
autorestart=true\n\
stdout_logfile=/dev/stdout\n\
stdout_logfile_maxbytes=0\n\
stderr_logfile=/dev/stderr\n\
stderr_logfile_maxbytes=0\n\
environment=PATH="/opt/venv/bin:%(ENV_PATH)s"' > /etc/supervisor/conf.d/supervisord.conf

EXPOSE 80

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
