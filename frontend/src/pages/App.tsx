import { useState, useEffect } from 'react'
import ExerciseButton from "../components/exerciseButton";
import '../App.css'

/*
type Workout = {
  weight: number,
  sets: number,
  reps: number
};
*/

type Exercise = {
  id: string,
  name: string,
  //workouts: Workout[]
};

type Data = {
  name: string,
  exercises: Exercise[]
};

function App() {
  const [data, setData] = useState<Data | undefined>(undefined);

  async function fetchApi() {
    // TODO: Figure out environment variables
    const raw = await fetch("http://localhost:8080/app", {
      credentials: "include",
    });


    const json = await raw.json();
    setData(json);
  }

  useEffect(() => {
    fetchApi();
  }, []);

  return (
    <>
      {data ? (
        <>
          <div className="navBar">
            <h1>
              {data?.name + "'s Gym Journal"}
            </h1>
            <a href="http://localhost:8080/logout/google">logout</a>
          </div>
          <div className="exerciseButtonGrid">
            {data.exercises.map((e, k) => {
              return <ExerciseButton key={k} create={false} name={e.name} id={e.id} />
            })}
            <ExerciseButton create={true} fetchApi={fetchApi} />
          </div >
        </>

      ) : (
        <div>No data loaded :(</div>
      )
      }


    </>
  )
}

export default App
