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

  // TODO: Refetch when modal form submitted without refreshing page
  useEffect(() => {
    async function fetchApi() {
      // TODO: Figure out environment variables
      const raw = await fetch("http://localhost:8080/app", {
        credentials: "include",
      });


      const json = await raw.json();
      setData(json);
    }
    fetchApi();
  }, []);


  return (
    <>
      {data ? (
        <>
          <h1>
            {data?.name}
          </h1>
          <div className="exerciseButtonGrid">


            {data.exercises.map((e, k) => {
              return <ExerciseButton key={k} create={false} name={e.name} id={e.id} />
            })}
            <ExerciseButton create={true} />


          </div >
        </>

      ) : (
        <div>No data loaded :(</div>
      )}


    </>
  )
}

export default App
