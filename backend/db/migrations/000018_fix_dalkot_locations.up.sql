DELETE FROM dalkot_locations 
WHERE name NOT IN (
    'Jakarta Pusat', 
    'Jakarta Selatan', 
    'Jakarta Barat', 
    'Jakarta Utara', 
    'Jakarta Timur'
);
