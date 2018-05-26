install:
		touch ~/.ghorg
		cp .env ~/.ghorg
homebrew:
		cp .env-sample ~/.ghorg
uninstall:
		rm ~/.ghorg
