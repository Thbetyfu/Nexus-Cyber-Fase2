from flask import Flask, request, jsonify, render_template_string

app = Flask(__name__)

# Mock OJK Investor Data Base
INVESTORS = {
    "inv_001": {"name": "Budi Santoso", "balance": "Rp 5.000.000"},
    "inv_002": {"name": "Siti Aisyah", "balance": "Rp 12.500.000"}
}

# State Global untuk Efek Defacement
website_state = {
    "is_defaced": False,
    "deface_message": ""
}

# UI HTML Halaman OJK Normal Terlindungi
HTML_NORMAL = """
<!DOCTYPE html>
<html>
<head>
    <title>Sistem Data Terpadu OJK (Mockup)</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f6f9; color: #333; margin: 0; padding: 50px; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 10px 15px rgba(0,0,0,0.05); border-top: 6px solid #2563eb; }
        h1 { color: #1e3a8a; margin-top: 0; }
        .secure-badge { display: inline-block; background: #dcfce7; color: #166534; padding: 6px 12px; border-radius: 20px; font-size: 13px; font-weight: bold; margin-bottom: 20px; }
        .footer { margin-top: 30px; font-size: 14px; color: #666; border-top: 1px solid #eee; padding-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>OJK Central Investor Portal v1.0</h1>
        <span class="secure-badge">✔️ SECURE CONNECTION ESTABLISHED</span>
        <p>Selamat datang di sistem data terpusat. Gateway aman beroperasi secara normal.</p>
        <div class="footer">
            <i>Status Server: Online | Total Investor Aktif: 2 | Server: Jakarta</i>
        </div>
    </div>
</body>
</html>
"""

# UI HTML Saat Di-*Hack* (Defacement)
HTML_DEFACED = """
<!DOCTYPE html>
<html>
<head>
    <title>HACKED !!!</title>
    <style>
        body { background-color: black; color: red; font-family: 'Courier New', Courier, monospace; text-align: center; padding-top: 20vh; margin: 0; overflow: hidden; }
        h1 { font-size: 6rem; text-shadow: 0 0 15px red; margin: 0; animation: glitch 1s linear infinite; }
        p { font-size: 2rem; color: #2fff00; text-shadow: 0 0 5px #2fff00; }
        
        @keyframes glitch {
            2%, 64% { transform: translate(2px, 0) skew(0deg); }
            4%, 60% { transform: translate(-2px, 0) skew(0deg); }
            62% { transform: translate(0, 0) skew(5deg); }
        }
    </style>
</head>
<body>
    <h1>HACKED BY ANONYMOUS</h1>
    <p>{{ message }}</p>
    <p style="font-size: 1rem; color: #666; margin-top: 50px;">SECURITY IS AN ILLUSION.</p>
</body>
</html>
"""

@app.route('/', methods=['GET'])
def home():
    """Mengembalikan UI Web Utama, berubah 180 derajat jika di-deface."""
    if website_state["is_defaced"]:
        return render_template_string(HTML_DEFACED, message=website_state["deface_message"])
    return render_template_string(HTML_NORMAL)

@app.route('/api/deface', methods=['POST'])
def deface():
    """Endpoint super ceroboh. Memungkinkan manipulasi UI global via JSON!"""
    data = request.get_json(silent=True) or {}
    new_title = data.get("new_title", "YOU HAVE BEEN PWNED!")
    
    website_state["is_defaced"] = True
    website_state["deface_message"] = new_title
    print(f"[!!!] FATAL: ROOT PAGE HAS BEEN DEFACED BY: {new_title}")
    
    return jsonify({"status": "success", "message": f"Website successfully defaced with message: {new_title}"}), 200

@app.route('/api/reset', methods=['GET'])
def reset():
    """Mengembalikan state website OJK ke normal."""
    website_state["is_defaced"] = False
    website_state["deface_message"] = ""
    print("[+] State Web Dipulihkan oleh Tim IT.")
    return jsonify({"status": "success", "message": "Website restored to normal"}), 200

@app.route('/api/investors', methods=['GET'])
def get_investors():
    """Endpoint rentan SQL Injeksi."""
    inv_id = request.args.get('id')
    
    if inv_id and ("OR 1=1" in inv_id.upper() or "'" in inv_id):
        print(f"[!] SQL INJECTION ATTACK DETECTED IN BACKEND: {inv_id}")
        return jsonify({
            "status": "success",
            "message": "DATA LEAKED VIA SQL INJECTION",
            "data": INVESTORS
        }), 200
        
    if inv_id in INVESTORS:
        return jsonify({"status": "success", "data": INVESTORS[inv_id]}), 200
        
    return jsonify({"status": "error", "message": "Portal Data Investor OJK", "data": "Tentukan ID investor"}), 400

if __name__ == '__main__':
    print("[+] OJK Mockup Server menyala di Port 3001")
    # Menonaktifkan debug agar tidak print 2x stack trace, tapi tetap hot-reload (jika dari cmd).
    app.run(host='0.0.0.0', port=3001)
