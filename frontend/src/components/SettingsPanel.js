// src/components/SettingsPanel.js
import React, { useState } from 'react';

const SettingsPanel = ({ onSettingsChange }) => {
  const [theme, setTheme] = useState("vs-dark");
  const [fontSize, setFontSize] = useState(14);

  const handleThemeChange = (e) => {
    const newTheme = e.target.value;
    setTheme(newTheme);
    onSettingsChange({ theme: newTheme, fontSize });
  };

  const handleFontSizeChange = (e) => {
    const newFontSize = Number(e.target.value);
    setFontSize(newFontSize);
    onSettingsChange({ theme, fontSize: newFontSize });
  };

  return (
    <div className="settings-panel">
      <label>
        Theme:
        <select value={theme} onChange={handleThemeChange}>
          <option value="vs-dark">Dark</option>
          <option value="vs-light">Light</option>
        </select>
      </label>
      <label>
        Font Size:
        <input type="number" value={fontSize} onChange={handleFontSizeChange} />
      </label>
    </div>
  );
};

export default SettingsPanel;
