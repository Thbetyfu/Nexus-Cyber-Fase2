from http.server import HTTPServer, BaseHTTPRequestHandler
import sys
class Mock(BaseHTTPRequestHandler):
    def do_GET(self):
        print(f"[BACKEND {sys.argv[1]}] Incoming hit!")
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b'Welcome to Nexus Protected Asset ' + sys.argv[1].encode())
port = int(sys.argv[1])
httpd = HTTPServer(('localhost', port), Mock)
print(f"Python Mock Active on {port}")
httpd.serve_forever()
