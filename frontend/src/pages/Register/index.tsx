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

	return (
		<>
			<label htmlFor="">
				<input
					type="text"
					name=""
					id=""
					onChange={(e: ChangeEvent<HTMLInputElement>) =>
						setUser({
							username: e.currentTarget.value,
						})
					}
				/>
			</label>
			<label htmlFor="">
				<input
					type="text"
					name=""
					id=""
					onChange={(e: ChangeEvent<HTMLInputElement>) =>
						setUser({
							email: '',
							fullName: '',
							password: '',
							username: e.currentTarget.value,
						})
					}
				/>
			</label>
			<label htmlFor="">
				<input
					type="email"
					name=""
					id=""
					onChange={(e: ChangeEvent<HTMLInputElement>) =>
						setEmail(e.currentTarget.value)
					}
				/>
			</label>
			<label htmlFor="">
				<input
					type="text"
					name=""
					id=""
					onChange={(e: ChangeEvent<HTMLInputElement>) =>
						setUsername(e.currentTarget.value)
					}
				/>
			</label>
			<label htmlFor="">
				<input
					type="password"
					name=""
					id=""
					onChange={(e: ChangeEvent<HTMLInputElement>) =>
						setPassword(e.currentTarget.value)
					}
				/>
			</label>
			<button type="submit">Register</button>
		</>
	)
}
