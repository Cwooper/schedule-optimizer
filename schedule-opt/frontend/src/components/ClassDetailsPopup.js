// ClassDetailsPopup.js
import React from "react";

const ClassDetailsPopup = ({ course, onClose }) => {
    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
                <div className="mt-3 text-center">
                    <h3 className="text-2xl leading-6 font-medium text-gray-900 pb-3 border-b border-gray-300">
                        {course.subject}
                    </h3>
                    <button
                        onClick={onClose}
                        className="absolute top-0 right-0 mt-2 mr-4 text-gray-400 hover:text-gray-600 text-3xl">
                        &times;
                    </button>
                    <div className=" px-7 py-3">
                        {Object.entries(course).map(([key, value]) => (
                            <p key={key} className="text-sm text-gray-600">
                                <strong>{key}:</strong> {value}
                            </p>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ClassDetailsPopup;
