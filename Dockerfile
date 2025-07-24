FROM python:3.11-slim

WORKDIR /app

# Create a simple HTTP server script
RUN cat <<'EOF' > server.py
#!/usr/bin/env python3
import http.server
import socketserver
import os
import socket
import json
import sys
from datetime import datetime

class DebugHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()
        
        # Collect debug information
        debug_info = {
            "timestamp": datetime.now().isoformat(),
            "request": {
                "path": self.path,
                "headers": dict(self.headers),
                "client_address": self.client_address[0],
                "client_port": self.client_address[1]
            },
            "host": {
                "hostname": socket.gethostname(),
                "fqdn": socket.getfqdn(),
                "ip_addresses": self.get_ip_addresses()
            },
            "environment_variables": dict(os.environ),
            "python_version": os.sys.version
        }
        
        response = json.dumps(debug_info, indent=2, ensure_ascii=False)
        self.wfile.write(response.encode('utf-8'))
    
    def get_ip_addresses(self):
        """Get all IP addresses of the host"""
        ips = []
        try:
            hostname = socket.gethostname()
            for addr_info in socket.getaddrinfo(hostname, None):
                ip = addr_info[4][0]
                if ip not in ips:
                    ips.append(ip)
        except:
            pass
        return ips
    
    def log_message(self, format, *args):
        """Override to include timestamp in logs"""
        print(f"[{datetime.now().isoformat()}] {format % args}")

if __name__ == "__main__":
    # Get port from command line argument or environment variable
    if len(sys.argv) > 1:
        PORT = int(sys.argv[1])
    else:
        PORT = int(os.environ.get('PORT', 9876))
    
    with socketserver.TCPServer(("", PORT), DebugHTTPRequestHandler) as httpd:
        print(f"Debug HTTP server starting on port {PORT}")
        print(f"Access at http://localhost:{PORT}")
        print("Press Ctrl-C to stop")
        httpd.serve_forever()
EOF

# Make the script executable
RUN chmod +x server.py

# Expose the default port
EXPOSE 9876

# Run the server with default port (can be overridden by CMD)
ENTRYPOINT ["python", "server.py"]
CMD ["9876"]