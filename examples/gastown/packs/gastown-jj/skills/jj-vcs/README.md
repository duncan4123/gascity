# Opinionated Jujutsu (jj) skill for Agents

I am an avid user of [Jujutsu](https://github.com/martinvonz/jj) VCS for couple of years. You might have read my [blog post](https://thisalex.com/posts/2025-04-20/) about it. This is my opinionated skill for using it with Agents.

## Key features

- Agents know how to write [good commit messages](https://cbea.ms/git-commit/) and how to chose a proper [gitmoji](https://gitmoji.dev/).
- Agents never push to remote repositories — _I always verify the changes before publication and so should you_.
- Agents are discouraged from deleting their failed attempts. Those are kept in separate branches for future reference — _because jj makes it really easy_.
- Agents know how to distribute hunks of code across multiple commits if you ask them to.
- Agents are not added as co-authors — _I as a person am fully responsible for the changes and do not plan to shift the blame to the tool_.

## Installation

> **Quick start:** Give your agent this page URL and ask it to install the jj-vcs skill. It will read the instructions below and handle the rest.

_(you can use `git clone` instead of `jj git clone --colocate` in the following commands if you prefer)_

### Claude Code

```bash
jj git clone --colocate https://codeberg.org/thisalex/jj-skill.git ~/.claude/skills/jj-vcs
```

### Zed / Codex / Pi

```bash
mkdir -p ~/.agents/skills
jj git clone --colocate https://codeberg.org/thisalex/jj-skill.git ~/.agents/skills/jj-vcs
```

### Hermes

```bash
mkdir -p ~/.hermes/skills
jj git clone --colocate https://codeberg.org/thisalex/jj-skill.git ~/.hermes/skills/jj-vcs
```

### Other agents

The skill follows the [Agent Skills](https://agentskills.io) format. `SKILL.md` is the entry point; other `.md` files are referenced from it when needed. Check your agent's documentation for the correct skills directory and clone there.

### Recommended: add a reinforcement to your global instructions

The skill works on its own, but agents may skip loading it if they think they already know jj. Adding a line to your global instructions (e.g. `CLAUDE.md`, `.cursor/rules`, etc.) ensures the skill is always loaded:

```
Always load the jj-vcs skill when starting any programming task.
Do not rely on general knowledge for jj workflow or message format —
the skill contains the required policies.
```

## Support me

[☕ Buy me a coffee](https://www.buymeacoffee.com/thisalex)

p.s. This readme is written by human. All of the mdashes are inserted manually with ❤️.
