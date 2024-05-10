# app.py

from flask import Flask, jsonify, request, render_template, redirect, url_for, send_from_directory

from backend.schedule_generator import generate_schedules, get_courses

app = Flask(__name__, template_folder="frontend", static_folder="/schedule-optimizer/frontend")


@app.route('/')
def index():
    return redirect(url_for('schedule_optimizer'))

@app.route('/schedule-optimizer/')
def schedule_optimizer():
    return render_template("index.html")

# Define route to serve static files
@app.route('/schedule-optimizer/frontend/<path:filename>')
def serve_static(filename):
    return send_from_directory('frontend', filename)

@app.route('/schedule-optimizer', methods=['POST'])
def generate_response():
    data = request.json
    print(data)

    # Create my "response" object
    response = {
        "schedules": [],
        "warnings": [],
        "errors": []
    }

    course_dict = get_courses(data["courses"], data["term"])
    if course_dict["warnings"] != []:
        response["warnings"].append(course_dict["warnings"])
    if course_dict["errors"] != []:
        response["errors"].append(course_dict["errors"])

    if response["errors"] != []:
        return jsonify(response) # Return the errors if the program failed
    
    courses = course_dict["courses"]
    schedules_response = generate_schedules(courses,
                                   min_schedule_size=int(data["min"]),
                                   max_schedule_size=int(data["max"]))

    response["schedules"]= [schedule.to_dict() for schedule in schedules_response["schedules"]]
    if schedules_response["warnings"] != []:
        response["warnings"].append(schedules_response["warnings"])
    if schedules_response["errors"] != []:
        response["errors"].append(schedules_response["errors"])

    return jsonify(response)

if __name__ == "__main__":
    app.run(debug=True)
