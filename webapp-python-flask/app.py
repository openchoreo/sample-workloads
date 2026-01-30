from flask import Flask, render_template, jsonify
import os

app = Flask(__name__)

@app.route('/')
def home():
    return render_template('index.html')

@app.route('/api/health')
def health():
    return jsonify({
        'status': 'healthy',
        'app': 'Python Flask Web App',
        'version': '1.0.0'
    })

@app.route('/api/info')
def info():
    return jsonify({
        'framework': 'Flask',
        'language': 'Python',
        'environment': os.getenv('FLASK_ENV', 'production')
    })

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
