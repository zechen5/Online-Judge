# Frontend-Only Subset

This folder is a standalone frontend extract for the Online Judge project.

## What it keeps

- AlgoJudge branding and overall UI direction
- Home, Problems, and Status views
- Local login/register demo
- Mock submission flow stored in `localStorage`

## What it removes

- Go backend dependency
- Real `/api/v1/*` calls
- Database and judge integration

## How to use

Open `index.html` in a browser, or serve this folder with any static server.

Demo accounts:

- `demo_student` / `password`
- `demo_admin` / `password`
