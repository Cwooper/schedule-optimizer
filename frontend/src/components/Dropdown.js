import React from "react";

const Dropdown = ({ label, value, options, onChange, className = "" }) => {
    return (
        <div className={`form-control ${className}`}>
            <label className="label">
                <span className="label-text text-lightgray">{label}</span>
            </label>
            <select
                className="select select-bordered bg-white border-lightgray text-darkgray w-full"
                value={value}
                onChange={onChange}>
                {options.map((option) => {
                    if (typeof option === "object" && option !== null) {
                        return (
                            <option key={option.value} value={option.value}>
                                {option.label}
                            </option>
                        );
                    } else {
                        return (
                            <option key={option} value={option}>
                                {option}
                            </option>
                        );
                    }
                })}
            </select>
        </div>
    );
};

export default Dropdown;
