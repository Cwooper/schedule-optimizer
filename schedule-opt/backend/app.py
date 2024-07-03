import os
from flask import Flask, jsonify, request, send_from_directory
from flask_cors import CORS

import sys
# Add the parent directory to the Python path
parent_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.append(parent_dir)

from backend.schedule_generator import generate_schedules, get_courses

app = Flask(__name__, static_folder='../frontend/build')
CORS(app)  # Enable CORS for all routes

@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def serve(path):
    if path != "" and os.path.exists(app.static_folder + '/' + path):
        return send_from_directory(app.static_folder, path)
    else:
        return send_from_directory(app.static_folder, 'index.html')

@app.route('/api/schedule-optimizer', methods=['POST'])
def generate_response():
    data = request.json
    forced_courses = data['force']
    course_names = data["courses"]
    minimum = int(data['min'])
    maximum = int(data['max'])
    term = data['term']

    # Why waste unnecessary CPU time
    if len(forced_courses) > minimum:
        minimum = len(forced_courses)

    # Create my "response" object
    response = {
        "schedules": [],
        "warnings": [],
        "errors": []
    }

    course_dict = get_courses(course_names, term)
    if course_dict["warnings"]:
        response["warnings"].extend(course_dict["warnings"])
    if course_dict["errors"]:
        response["errors"].extend(course_dict["errors"])

    if response["errors"]:
        return jsonify(response)  # Return the errors if the program failed
    
    courses = course_dict["courses"]
    schedules_response = generate_schedules(courses, minimum, maximum, forced_courses)

    schedules = schedules_response["schedules"]

    response["schedules"] = [schedule.to_dict() for schedule in schedules]
    if schedules_response["warnings"]:
        response["warnings"].extend(schedules_response["warnings"])
    if schedules_response["errors"]:
        response["errors"].extend(schedules_response["errors"])

    return jsonify(response)

if __name__ == "__main__":
    app.run(debug=True)