install:
		touch ~/.ghorg
		cp .env ~/.ghorg
homebrew:
		touch ~/.ghorg
		cp .env-sample ~/.ghorg
uninstall:
		rm ~/.ghorg
