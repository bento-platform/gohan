flask db upgrade;
# flask run --host=0.0.0.0;
gunicorn chord_drs.app:application -w 1 --threads $(expr 2 \* $(nproc --all) + 1) -b 0.0.0.0:${INTERNAL_PORT}
