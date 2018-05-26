install:
		touch ~/.ghorg
		cp .env ~/.ghorg
homebrew:
		touch ${HOME}/.ghorg
		cp .env-sample ${HOME}/.ghorg
uninstall:
		rm ~/.ghorg
