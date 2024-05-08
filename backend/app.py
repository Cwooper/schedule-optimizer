# backend/app.py

from flask import Flask, jsonify, request

from models import Course, Schedule

from schedule_generator import generate_schedules, clean_course_names

app = Flask(__name__)

@app.route('/schedule-optimizer', methods =['POST'])
def generate_schedules():
    data = request.json
    print(data)
    courses = clean_course_names(courses)
    schedules = generate_schedules(data)
    return jsonify(schedules)

if __name__ == "__main__":
    app.run(debug=True)
