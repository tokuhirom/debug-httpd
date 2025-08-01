#!/usr/bin/env python3
import http.server
import socketserver
import os
import socket
import json
import sys
from datetime import datetime
from collections import deque
import threading

# Global access log storage (thread-safe)
access_logs = deque(maxlen=100)
access_logs_lock = threading.Lock()

class DebugHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        # Log access for all requests
        self.log_access()
        
        # Handle /ping endpoint
        if self.path == '/ping':
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b'pong')
            return
        
        # Handle /logs endpoint
        if self.path == '/logs':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            with access_logs_lock:
                logs = list(access_logs)
            
            response = json.dumps(logs, indent=2, ensure_ascii=False)
            self.wfile.write(response.encode('utf-8'))
            return
        
        # Default handler for all other paths
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
    
    def do_POST(self):
        self.log_access()
        self.send_error(405, "Method Not Allowed")
    
    def do_PUT(self):
        self.log_access()
        self.send_error(405, "Method Not Allowed")
    
    def do_DELETE(self):
        self.log_access()
        self.send_error(405, "Method Not Allowed")
    
    def do_HEAD(self):
        self.log_access()
        self.send_error(405, "Method Not Allowed")
    
    def log_access(self):
        """Log access information"""
        log_entry = {
            "timestamp": datetime.now().isoformat(),
            "method": self.command,
            "path": self.path,
            "client_address": self.client_address[0],
            "client_port": self.client_address[1],
            "user_agent": self.headers.get('User-Agent', ''),
            "referer": self.headers.get('Referer', ''),
            "host": self.headers.get('Host', '')
        }
        
        with access_logs_lock:
            access_logs.append(log_entry)
    
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