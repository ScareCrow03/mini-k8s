import os
from flask import Flask, request, jsonify
import func
import json

host = "0.0.0.0"
port = 10000
app = Flask(__name__)
@app.route('/', methods=['GET'])
def handle_get():
    print("Hello, world!")
    return 'Hello, World!'
@app.route('/', methods=['POST'])
def handle_post():
    data = json.loads(request.data)
    res = func.handle(data)
    return jsonify(res)

if __name__ == '__main__':
    app.run(host=host, port=port)