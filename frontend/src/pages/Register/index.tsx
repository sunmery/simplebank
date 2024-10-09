import type {ChangeEvent} from 'react'
import {useState} from 'react'

const createUser = async (
	username: string,
	fullName: string,
	email: string,
	password: string,
) => {
	try {
		const res = await fetch('http://localhost:8080/users', {
			method: 'PUT',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({username, fullName, email, password}),
		})
		const data: LoginResponse = await res.json()
		return data
	} catch (error) {
		if (error instanceof Error) {
			throw new Error(error.message)
		}
		throw new Error(String(error))
	}
}
interface User {
	username: string
	fullName: string
	email: string
	password: string
}

/**
 * @returns JSXElement
 */
export default function Register() {
	const [user, setUser] = useState<User>({
		username: '',
		fullName: '',
		email: '',
		password: '',
	})

	const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
		const {name, value} = e.target
		setUser((prevState) => ({
			...prevState,
			[name]: value,
		}))
	}

	return (
		<>
			<label htmlFor="">
				<input
					type="text"
					name="username"
					id=""
					value={user.username}
					onChange={handleInputChange}
				/>
			</label>
			<label htmlFor="">
				<input
					type="text"
					name="fullName"
					id=""
					onChange={handleInputChange}
				/>
			</label>
			<label htmlFor="">
				<input
					type="email"
					name="email"
					id=""
					onChange={handleInputChange}
				/>
			</label>

			<label htmlFor="">
				<input
					type="password"
					name="password"
					id=""
					onChange={handleInputChange}
				/>
			</label>
			<label htmlFor="">
				重复密码:
				<input
					type="password"
					name="repPassword"
					id=""
				/>
			</label>
			<button type="submit">Register</button>
		</>
	)
}
