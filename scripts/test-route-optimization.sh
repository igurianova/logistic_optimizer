#!/bin/bash
curl --location 'http://localhost:8080/vrp-calculator' \
--form 'cargoCapacity=3.5' \
--form 'matrix=@"assets/matrix.csv"' \
--form 'warehouse-product-capacity=@"assets/warehouse-product-capacity.csv"'