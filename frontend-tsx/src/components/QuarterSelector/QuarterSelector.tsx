import React from 'react';

interface QuarterSelectorProps {
  quarter: string;
  year: string;
  minCredits: string;
  maxCredits: string;
  onUpdate: (field: string, value: string) => void;
}

const QuarterSelector: React.FC<QuarterSelectorProps> = ({
  quarter,
  year,
  minCredits,
  maxCredits,
  onUpdate,
}) => {
  return (
    <div className="flex gap-4 p-4">
      <div className="flex flex-col gap-2">
        <label className="text-sm font-medium">Quarter:</label>
        <div className="relative">
          <select
            value={quarter}
            onChange={(e) => onUpdate('quarter', e.target.value)}
            className="pl-3 pr-8 py-2 border rounded-md appearance-none bg-white"
          >
            <option value="">Select Quarter</option>
            <option value="fall">Fall</option>
            <option value="winter">Winter</option>
            <option value="spring">Spring</option>
            <option value="summer">Summer</option>
          </select>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <label className="text-sm font-medium">Year:</label>
        <div className="relative">
          <select
            value={year}
            onChange={(e) => onUpdate('year', e.target.value)}
            className="pl-3 pr-8 py-2 border rounded-md appearance-none bg-white"
          >
            <option value="">Select Year</option>
            {[2024, 2025].map(year => (
              <option key={year} value={year}>{year}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <label className="text-sm font-medium">Min:</label>
        <div className="relative">
          <select
            value={minCredits}
            onChange={(e) => onUpdate('minCredits', e.target.value)}
            className="pl-3 pr-8 py-2 border rounded-md appearance-none bg-white"
          >
            <option value="">Min Credits</option>
            {Array.from({ length: 8 }, (_, i) => i + 1).map(num => (
              <option key={num} value={num}>{num}</option>
            ))}
          </select>
        </div>
      </div>

      <div className="flex flex-col gap-2">
        <label className="text-sm font-medium">Max:</label>
        <div className="relative">
          <select
            value={maxCredits}
            onChange={(e) => onUpdate('maxCredits', e.target.value)}
            className="pl-3 pr-8 py-2 border rounded-md appearance-none bg-white"
          >
            <option value="">Max Credits</option>
            {Array.from({ length: 8 }, (_, i) => i + 1).map(num => (
              <option key={num} value={num}>{num}</option>
            ))}
          </select>
        </div>
      </div>
    </div>
  );
};

export default QuarterSelector;