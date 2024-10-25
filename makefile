check-links:
	npm install -g broken-link-checker
	blc http://localhost:1313 -ro | grep BROKEN


