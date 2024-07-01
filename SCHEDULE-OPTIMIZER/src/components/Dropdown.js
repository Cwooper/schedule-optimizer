import React from "react";

const Dropdown = ({ label, value, options, onChange }) => {
    return (
        <div className="form-control w-36 max-w-xs">
            <label className="label">
                <span className="label-text text-lightgray">{label}</span>
            </label>
            <select
                className="select select-bordered bg-white border-lightgray text-darkgray"
                value={value}
                onChange={onChange}>
                {options.map((option) => (
                    <option key={option} value={option}>
                        {option}
                    </option>
                ))}
            </select>
        </div>
    );
};

export default Dropdown;
