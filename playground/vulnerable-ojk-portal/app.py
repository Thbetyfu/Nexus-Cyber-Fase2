from flask import Flask, render_template, jsonify, request

app = Flask(__name__)

# State Global untuk Visual Demo
state = {
    "is_defaced": False,
    "deface_message": "HACKED BY APT-X: YOU LOSER!"
}

@app.route('/')
def index():
    if state["is_defaced"]:
        return render_template('hacked.html', message=state["deface_message"])
    return render_template('index.html')

@app.route('/api/status', methods=['GET'])
def get_status():
    """Endpoint untuk polling frontend."""
    return jsonify({
        "status": "online",
        "is_defaced": state["is_defaced"]
    }), 200

@app.route('/api/system/override', methods=['POST'])
def override():
    """Endpoint rentan tanpa otentikasi untuk simulasi defacement."""
    data = request.json or {}
    command = data.get("command")
    
    if command == "deface":
        state["is_defaced"] = True
        print("[!!!] SYSTEM OVERRIDE: Website telah di-deface!")
        return jsonify({"status": "success", "message": "System Hijacked"}), 200
    
    return jsonify({"status": "error", "message": "Invalid Command"}), 400

@app.route('/api/system/restore', methods=['POST'])
def restore():
    """Endpoint untuk pemulihan demo."""
    state["is_defaced"] = False
    print("[+] SYSTEM RESTORE: Website kembali normal.")
    return jsonify({"status": "success", "message": "System Restored"}), 200

@app.route('/api/investors', methods=['GET'])
def get_investors():
    """Legacy vulnerable endpoint (SQLi simulator)."""
    return jsonify({"status": "success", "data": "Mockup Data Protected"}), 200

if __name__ == '__main__':
    print("[+] FULLSTACK OJK PORTAL menyala di Port 3001")
    app.run(host='0.0.0.0', port=3001)
