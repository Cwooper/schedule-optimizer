import React from "react";

const CourseList = ({ courses, onRemoveCourse, onToggleForce }) => {
    return (
        <div className="join join-vertical border-[1px] border-100 w-full ">
            {courses.length === 0 ? (
                <div className="card bg-white w-full join-item">
                    <div className="card-body p-4">
                        <div className="flex justify-center items-center">
                            <div className="font-bold text-lightgray">
                                Select courses above
                            </div>
                        </div>
                    </div>
                </div>
            ) : (
                courses.map((course) => (
                    <div
                        key={course.name}
                        className="card bg-white w-full join-item">
                        <div className="card-body p-4">
                            <div className="flex justify-between items-center">
                                <div className="font-bold">{course.name}</div>
                                <div>
                                    <button
                                        className="btn btn-outline mr-2"
                                        onClick={() =>
                                            onToggleForce(course.name)
                                        }>
                                        {course.forced ? "Unforce" : "Force"}
                                    </button>
                                    <button
                                        className="btn btn-outline"
                                        onClick={() =>
                                            onRemoveCourse(course.name)
                                        }>
                                        –
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                ))
            )}
        </div>
    );
};

export default CourseList;
