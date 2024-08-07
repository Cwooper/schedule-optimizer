// ClassDetailsPopup.js
import React from "react";

const ClassDetailsPopup = ({ course, onClose }) => {
    const renderValue = (key, value) => {
        if (Array.isArray(value)) {
            return (
                <ul className="list-disc pl-4 leading-tight">
                    {value.map((item, index) => (
                        <li key={index} className="mb-1">
                            {Object.entries(item).map(([subKey, subValue]) => (
                                <div key={subKey} className="leading-tight">
                                    <strong>{subKey}:</strong>{" "}
                                    {renderValue(subKey, subValue)}
                                </div>
                            ))}
                        </li>
                    ))}
                </ul>
            );
        } else if (typeof value === "object" && value !== null) {
            return (
                <span className="leading-tight">{JSON.stringify(value)}</span>
            );
        } else if (typeof value === "number" && !Number.isInteger(value)) {
            return <span className="leading-tight">{value.toFixed(2)}</span>;
        } else {
            return <span className="leading-tight">{value.toString()}</span>;
        }
    };

    return (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
            <div className="relative top-20 mx-auto p-4 border lg:w-3/12 sm:w-8/12 shadow-lg rounded-md bg-white">
                <div className="mt-2 text-center">
                    <h3 className="text-xl leading-10 font-medium text-gray-900 pb-2 border-b border-gray-300">
                        {course.Subject}
                    </h3>
                    <button
                        onClick={onClose}
                        className="absolute top-0 right-0 mt-1 mr-2 text-gray-400 hover:text-gray-600 text-2xl">
                        &times;
                    </button>
                    <div className="px-4 py-2 text-left max-h-96 overflow-y-auto leading-tight text-sm">
                        {Object.entries(course).map(([key, value]) => (
                            <div key={key} className="mb-1">
                                <strong>{key}:</strong>
                                <div className="ml-1">
                                    {renderValue(key, value)}
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ClassDetailsPopup;
