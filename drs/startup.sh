mkdir -p /drs/chord_drs/data/obj;
mkdir -p /drs/chord_drs/data/db;

flask db upgrade;
flask run --host=0.0.0.0;