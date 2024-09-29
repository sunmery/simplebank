import {createRootRoute, Link, Outlet} from '@tanstack/react-router'

export const Route = createRootRoute({
	component: () => (
		<>
			<div>
				<Link to="/">Index</Link>
				<Link to="/login">Login</Link>
			</div>
			<Outlet />
		</>
	),
})
